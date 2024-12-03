import postgres from "postgres";

import { Category, Image, LocalImageInfo, TagEntry, TagMap } from "../types/drawref.js";
import urlJoin from "url-join";
import { allowLocalImages, myURL, uploadUrlPrefix } from "../config/env.js";

function tagMapToDbTags(tags: TagMap): string[] {
  if (tags === undefined) {
    return [];
  }

  var tagsForDb: Array<string> = [];
  Object.entries(tags).forEach(([key, values]) => {
    tagsForDb = tagsForDb.concat(values.map((value) => `${key} ${value}`));
  });
  return tagsForDb;
}

function imageUrl(row: postgres.Row): string {
  return row.is_local
    ? urlJoin(myURL, "local/img", String(row.image_id))
    : row.path
      ? urlJoin(uploadUrlPrefix, row.path)
      : row.external_url;
}

let db: Database;

class Database {
  private sql: postgres.Sql;

  constructor(url: string, options?: postgres.Options<{}>) {
    this.sql = postgres(url, options);
  }

  // Categories
  //

  async addCategory(id: string, name: string, cover: number, tags: Array<TagEntry>): Promise<string | undefined> {
    if (tags.filter((info) => info.id.includes(" ")).length > 0) {
      return "Error: tag IDs cannot contain spaces";
    }

    var newId: string = "";
    try {
      const row = await this.sql`
        insert into categories
          (id, display_name, cover_image, tags)
        values
          (${id}, ${name}, ${cover === undefined ? -1 : cover}, ${JSON.stringify(tags)})
        returning id
      `;
      newId = row[0]?.id;
    } catch (error) {
      // error
      console.error(error);
      return;
    }
    return newId;
  }

  async editCategory(id: string, name: string, cover: number, tags: Array<TagEntry>): Promise<string | undefined> {
    if (tags.filter((info) => info.id.includes(" ")).length > 0) {
      return "Error: tag IDs cannot contain spaces";
    }

    try {
      await this.sql`
        update categories
        set
          display_name = ${name},
          cover_image = ${cover === undefined ? -1 : cover},
          tags = ${JSON.stringify(tags)}
        where id = ${id}
      `;
    } catch (error) {
      // error
      console.error(error);
      return `Error: ${error}`;
    }
    return;
  }

  // note: images in the category won't be deleted, just unlinked from the category
  async deleteCategory(id: string) {
    try {
      await this.sql`
      delete from image_tags
      where
        category_id = ${id}
      `;
    } catch (error) {
      // error
      console.error("could not delete category images:", error);
      return `Images error: ${error}`;
    }

    try {
      await this.sql`
        delete from categories
        where
          id = ${id}
      `;
    } catch (error) {
      // error
      console.error(error);
      return `Error: ${error}`;
    }
    return;
  }

  async getCategories() {
    var categories: Category[] = [];

    //TODO: make this one query, using joins etc
    const rows = await this.sql`
      select *
      from categories
      order by position asc, id desc
    `;
    for (const row of rows) {
      var cat: Category = {
        id: row.id,
        name: row.display_name,
        tags: JSON.parse(row.tags || "[]"),
        cover: "",
      };
      if (row.cover_image != -1) {
        cat.cover_id = row.cover_image;
        // load image data and apply to cat
        const ciRows = await this.sql`
          select path, external_url, is_local
          from images
          where id = ${row.cover_image}
        `;
        cat.cover = imageUrl(ciRows[0]);
      }

      categories.push(cat);
    }

    return categories;
  }

  async reorderCategories(ids: string[]): Promise<string> {
    try {
      // ensure that all category ids are in the `ids` list
      const rows = await this.sql`
        select id from categories
      `;
      const allCategoryIds: string[] = rows.map((row) => row.id);

      for (const id of ids) {
        let index = allCategoryIds.indexOf(id);
        if (index === -1) {
          throw "Unknown or duplicate category ID given";
        }
        allCategoryIds.splice(index, 1);
      }

      // append all un-ordered category IDs to the end of `ids`
      ids = ids.concat(allCategoryIds.sort());

      // reorder!
      const update_data = ids.map((id, index) => [id, index]);

      await this.sql`
        update categories
        set position = (update_data.position)::int
        from (values ${this.sql(update_data)}) as update_data (id, position)
        where categories.id = update_data.id
      `;
    } catch (error) {
      // error
      console.error("could not reorder categories:", error);
      return `${error}`;
    }

    return "";
  }

  // Images
  //

  async getLocalImagePath(image: number): Promise<LocalImageInfo | undefined> {
    if (!allowLocalImages) {
      console.error("Can't get local image path, local images are disabled");
      return;
    }

    try {
      const rows = await this.sql`
        select path, is_local
        from images
        where id = ${image}
      `;

      if (rows.length === 0 || !rows[0].is_local) {
        return;
      }

      const li: LocalImageInfo = {
        path: rows[0].path,
      };
      return li;
    } catch (error) {
      // error
      console.error(error);
    }
    return;
  }

  async addImage(
    path: string,
    external_url: string,
    author: string,
    author_url: string,
    is_local: boolean,
  ): Promise<number | undefined> {
    var newId: number;
    try {
      const row = await this.sql`
        insert into images
          (path, external_url, author, author_url, is_local)
        values
          (${path}, ${external_url}, ${author}, ${author_url}, ${is_local})
        returning id
      `;
      newId = row[0]?.id;
    } catch (error) {
      // error
      console.error(error);
      return;
    }
    return newId;
  }

  async addLocalImages(filenames: string[], author: string, author_url: string): Promise<number[] | undefined> {
    const insertedRows = filenames.map((fn) => {
      return {
        path: fn,
        author,
        author_url,
        is_local: true,
      };
    });
    try {
      const rows = await this.sql`
        insert into images ${this.sql(insertedRows)}
        returning id
      `;

      const ids: number[] = [];
      for (const row of rows) {
        ids.push(row.id);
      }
      return ids;
    } catch (error) {
      // error
      console.error(error);
    }

    return;
  }

  async addLocalImageImportPath(
    path: string,
    category_id: string,
    author: string,
    author_url: string,
    tags: TagMap,
  ): Promise<number | undefined> {
    //TODO: exit if path already exists in our db
    var existingId: string = "";
    try {
      const row = await this.sql`
        select id
        from local_image_imports
        where path=${path} and category_id=${category_id}
      `;
      existingId = row[0]?.id;
    } catch (error) {
      // error
      console.error(error);
    }

    if (existingId) {
      console.log("Local image import path already exists");
      return;
    }

    // no existing entry for path found, adding new one
    const tagsForDb = tagMapToDbTags(tags);

    var newId: number;
    try {
      const row = await this.sql`
        insert into local_image_imports
          (path, category_id, author, author_url, tags)
        values
          (${path}, ${category_id}, ${author}, ${author_url}, ${tagsForDb})
        returning id
      `;
      newId = row[0]?.id;
    } catch (error) {
      // error
      console.error(error);
      return;
    }
    return newId;
  }

  async deleteUnusedImages() {
    try {
      const rows = await this.sql`
        select id
        from images
        left outer join image_tags
          on images.id = image_tags.image_id
        where image_tags.image_id is null
          and images.id not in (select cover_image from categories)
        order by id
      `;

      const ids: number[] = [];
      for (const row of rows) {
        ids.push(row.id);
      }

      await this.sql`
        delete from images
        where
          id = any(${ids})
      `;
    } catch (error) {
      // error
      console.error(error);
    }
  }

  async addImageToCategory(category: string, image: number, tags: TagMap) {
    const tagsForDb = tagMapToDbTags(tags);

    try {
      await this.sql`
        insert into image_tags
          (category_id, image_id, tags)
        values
          (${category}, ${image}, ${tagsForDb})
      `;
    } catch (error) {
      // error
      console.error(error);
    }
  }

  async addImagesToCategory(category: string, images: number[], tags: TagMap) {
    const tagsForDb = tagMapToDbTags(tags);

    const insertedRows = images.map((img) => {
      return {
        category_id: category,
        image_id: img,
        tags: tagsForDb,
      };
    });

    try {
      await this.sql`
        insert into image_tags ${this.sql(insertedRows)}
      `;
    } catch (error) {
      // error
      console.error(error);
    }
  }

  async deleteImageFromCategory(category: string, image: number) {
    try {
      await this.sql`
        delete from image_tags
        where
          category_id = ${category} and
          image_id = ${image}
      `;
    } catch (error) {
      // error
      console.error(error);
    }
  }

  async getCategoryImages(category: string, page: number): Promise<Image[]> {
    var images: Image[] = [];

    const rows = await this.sql`
      select image_id, path, external_url, author, author_url, is_local, image_tags.tags
      from images
      inner join image_tags
        on images.id = image_tags.image_id
      where image_tags.category_id = ${category}
      order by image_tags.created_at desc
    `;
    for (const row of rows) {
      var img: Image = {
        id: row.image_id,
        path: imageUrl(row),
        author: row.author || "",
        author_url: row.author_url || "",
      };

      //TODO: turn tag list back into TagMap

      images.push(img);
    }

    return images;
  }

  async getSessionImages(category: string, tags: TagMap): Promise<Image[]> {
    var images: Image[] = [];
    const tagsForDb = tagMapToDbTags(tags);

    // most sessions won't be more than 30 images, so this should be fine
    var rows: postgres.RowList<postgres.Row[]>;
    if (tagsForDb.length === 0) {
      rows = await this.sql`
        select image_id, path, external_url, author, author_url, is_local
        from images
        inner join image_tags
          on images.id = image_tags.image_id
        where image_tags.category_id = ${category}
        order by RANDOM()
        limit 30
      `;
    } else {
      rows = await this.sql`
        select image_id, path, external_url, author, author_url, is_local
        from images
        inner join image_tags
          on images.id = image_tags.image_id
        where image_tags.category_id = ${category}
        and image_tags.tags @> ${tagsForDb}
        order by RANDOM()
        limit 30
      `;
    }
    for (const row of rows) {
      var img: Image = {
        id: row.image_id,
        path: imageUrl(row),
        author: row.author || "",
        author_url: row.author_url || "",
      };
      images.push(img);
    }

    return images;
  }

  async getSessionImageCount(category: string, tags: TagMap): Promise<number> {
    const tagsForDb = tagMapToDbTags(tags);

    // most sessions won't be more than 30 images, so this should be fine
    var rows: postgres.RowList<postgres.Row[]>;
    if (tagsForDb.length === 0) {
      rows = await this.sql`
        select count(*) as num
        from images
        inner join image_tags
          on images.id = image_tags.image_id
        where image_tags.category_id = ${category}
      `;
    } else {
      rows = await this.sql`
        select count(*) as num
        from images
        inner join image_tags
          on images.id = image_tags.image_id
        where image_tags.category_id = ${category}
        and image_tags.tags @> ${tagsForDb}
      `;
    }
    return rows[0].num;
  }

  async getImageSources() {
    var sources: string[][] = [];

    const rows = await this.sql`
      select distinct author, author_url
      from images
      order by author, author_url
    `;
    for (const row of rows) {
      if (row.author != "" || row.author_url != "") {
        sources.push([row.author, row.author_url]);
      }
    }

    return sources;
  }
}

export function initialiseDatabase(url: string, options?: postgres.Options<{}>) {
  if (db) {
    console.error("Database cannot be re-initialised:", url);
    return;
  }

  db = new Database(url, options);
}

export function useDatabase(): Database {
  if (!db) {
    console.error("Database is not initialised.");
  }

  return db;
}
