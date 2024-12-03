import express, { Request, Response } from "express";

import { needAdmin } from "../auth/authRequiredMiddleware.js";
import { parseSampleTags, sampleCategories, sampleImages } from "../sampleData.js";
import { useDatabase } from "../db/database.js";
import { uploadFile } from "../files/s3.js";

export const router = express.Router();

router.get("/", needAdmin, async (req: Request, res: Response) => {
  res.json({
    categories: sampleCategories,
    images: sampleImages.map((info) => {
      return {
        author: info.author,
        author_url: info.author_url,
        requirement: info.requirement,
        image_count: info.images.length,
      };
    }),
  });
});

router.post("/import", needAdmin, async (req: Request, res: Response) => {
  const categoriesToImport: string[] = req.body.categories || [];
  const imagesToImport: string[] = req.body.images || [];

  const db = useDatabase();

  for (const name of categoriesToImport) {
    const categoryData = sampleCategories.filter((info) => info.name === name).pop();
    if (!categoryData) {
      continue;
    }

    console.log("Importing sample category", categoryData.name);
    let cImage = -1;
    if (categoryData.cover) {
      const finalPath = `samples/covers/${categoryData.cover}`;
      cImage = (await db.addImage(finalPath, "", "", "", false)) || cImage;
      await uploadFile(`./samples/covers/${categoryData.cover}`, finalPath);
    }
    const cId = await db.addCategory(categoryData.id, categoryData.name, cImage, categoryData.tags || []);
    if (!cId) {
      continue;
    }

    for (const author of imagesToImport) {
      const imageProviderData = sampleImages.filter((info) => info.author === author).pop();
      if (!imageProviderData) {
        continue;
      }

      for (const imageList of imageProviderData.images) {
        if (imageList.category === categoryData.id) {
          for (const image of imageList.images) {
            const finalPath = `samples/${image.path}`;
            await uploadFile(`./samples/${image.path}`, finalPath);
            const iId = await db.addImage(finalPath, "", imageProviderData.author, imageProviderData.author_url, false);
            if (iId) {
              await db.addImageToCategory(cId, iId, parseSampleTags(image.tags || []));
            }
          }
        }
      }
    }
  }

  res.json({
    ok: true,
  });
});
