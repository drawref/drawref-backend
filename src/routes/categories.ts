import express, { Request, Response } from "express";

import { needAdmin } from "../auth/authRequiredMiddleware.js";
import { useDatabase } from "../db/database.js";
import { Category, TagEntry, TagMap } from "../types/drawref.js";

export const router = express.Router();

router.post("/:category/images/:image", needAdmin, async (req: Request, res: Response) => {
  const category = req.params.category || "";
  const image = parseInt(req.params.image || "-1");
  const tags: TagMap = req.body.tags;

  if (!category || image === -1) {
    res.status(400);
    res.json({
      error: "Required parameters not supplied.",
    });
    return;
  }

  const db = useDatabase();
  await db.addImageToCategory(category, image, tags);

  res.json({
    ok: true,
  });
});

router.delete("/:category/images/:image", needAdmin, async (req: Request, res: Response) => {
  const category = req.params.category || "";
  const image = parseInt(req.params.image || "-1");

  if (!category || image === -1) {
    res.status(400);
    res.json({
      error: "Required parameters not supplied.",
    });
    return;
  }

  const db = useDatabase();
  await db.deleteImageFromCategory(category, image);

  res.json({
    ok: true,
  });
});

router.get("/:category/images", async (req: Request, res: Response) => {
  const category = req.params.category || "";
  const page = parseInt(String(req.query.page) || "0");

  if (!category) {
    res.status(400);
    res.json({
      error: "Required parameters not supplied.",
    });
    return;
  }

  const db = useDatabase();

  const images = await db.getCategoryImages(category, page);

  res.send(images);
});

router.post("/", needAdmin, async (req: Request, res: Response) => {
  if (!req.body.id || !req.body.name) {
    res.status(400);
    res.json({
      error: "Required parameters not supplied.",
    });
    return;
  }

  const { id: cId, name: cName, cover: cImage } = req.body;
  const cTags: Array<TagEntry> = req.body.tags;

  if (cId === "" || cName === "") {
    res.status(400);
    res.json({
      error: "ID and name must be provided.",
    });
    return;
  }

  const db = useDatabase();
  const new_id = await db.addCategory(cId, cName, cImage, cTags);

  if (!new_id) {
    res.status(400);
    res.json({
      error: "Couldn't create category. Maybe it already exists.",
    });
    return;
  }

  res.json({
    id: new_id,
  });
});

router.put("/:id", async (req: Request, res: Response) => {
  var id = req.params.id || "";
  const { name: cName, cover: cImage } = req.body;
  const cTags: Array<TagEntry> = req.body.tags;

  if (cName === "") {
    res.status(400);
    res.json({
      error: "Name cannot be blank.",
    });
    return;
  }

  const db = useDatabase();
  const error = await db.editCategory(id, cName, cImage, cTags);

  if (error) {
    res.status(400);
    res.json({
      error: "Couldn't edit category.",
    });
    return;
  }

  res.json({
    id,
  });
});

router.delete("/:id", async (req: Request, res: Response) => {
  var id = req.params.id || "";

  const db = useDatabase();
  const error = await db.deleteCategory(id);

  if (error) {
    res.status(400);
    res.json({
      error: "Couldn't delete category.",
    });
    return;
  }

  res.json({
    ok: true,
  });
});

router.get("/", async (req: Request, res: Response) => {
  const db = useDatabase();
  const categories = await db.getCategories();

  res.json(categories);
});

router.get("/:id", async (req: Request, res: Response) => {
  var id = req.params.id || "";
  const db = useDatabase();
  const categories = await db.getCategories();
  var selectedCat: Category = categories.filter((cat) => cat.id === id)[0];

  res.json(selectedCat || { error: true, message: "Category not found" });
});
