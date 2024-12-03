import express, { Request, Response } from "express";
import { allowLocalImages, localImagesRootPath } from "../config/env.js";
import { useDatabase } from "../db/database.js";
import { join } from "path";

export const router = express.Router();

router.get("/img/:image", async (req: Request, res: Response) => {
  const image = parseInt(req.params.image || "-1");

  if (image === -1 || !allowLocalImages) {
    res.status(404);
    res.json({});
    return;
  }

  const db = useDatabase();

  const info = await db.getLocalImagePath(image);
  if (info === undefined) {
    res.status(404);
    res.json({});
    return;
  }

  res.sendFile(join(localImagesRootPath, info.path));
});
