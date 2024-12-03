import express, { Request, Response } from "express";
import { copyFileSync } from "fs";
import { join, relative } from "path";

import { needAdmin } from "../auth/authRequiredMiddleware.js";
import { useDatabase } from "../db/database.js";
import urlJoin from "url-join";
import { allowLocalImages, localImagesRootPath, uploadPathTmp, uploadUrlPrefix } from "../config/env.js";
import { uploadFile } from "../files/s3.js";
import { glob } from "glob";
import { isChildPath } from "../utilities/backstagePath.js";
import { TagMap } from "../types/drawref.js";

const supportedLocalImageExtensions: string[] = ["apng", "png", "avif", "gif", "jpg", "jpeg", "webp"];

export const router = express.Router();

router.post("/", needAdmin, async (req: Request, res: Response) => {
  const iPath = req.body.path || "";
  const iExternalUrl = req.body.external_url || "";
  const iAuthor = req.body.author || "";
  const iAuthorUrl = req.body.author_url || "";

  if ((!req.body.path && !req.body.external_url) || req.body.author === undefined) {
    res.status(400);
    res.json({
      error: "Required parameters not supplied.",
    });
    return;
  }

  // move image to the live directory
  const uploaded = uploadFile(join(uploadPathTmp, iPath), iPath);
  if (!uploaded) {
    res.status(400);
    res.json({
      error: "Could not upload file.",
    });
    return;
  }

  const db = useDatabase();
  const new_id = await db.addImage(iPath, iExternalUrl, iAuthor, iAuthorUrl, false);

  if (!new_id) {
    res.status(400);
    res.json({
      error: "Couldn't add image.",
    });
    return;
  }

  res.json({
    id: new_id,
    url: iExternalUrl ? iExternalUrl : urlJoin(uploadUrlPrefix, iPath),
  });
});

router.post("/local", needAdmin, async (req: Request, res: Response) => {
  if (!allowLocalImages) {
    res.status(400);
    res.json({
      error: "Local images are not allowed.",
    });
    return;
  }

  const iPath = req.body.folder || "";
  const iCategory = req.body.category || "";
  const iAuthor = req.body.author || "";
  const iAuthorUrl = req.body.author_url || "";
  const tags: TagMap = req.body.tags;

  if (!iPath || !iCategory || req.body.author === undefined) {
    res.status(400);
    res.json({
      error: "Required parameters not supplied.",
    });
    return;
  }

  if (!isChildPath(localImagesRootPath, iPath)) {
    res.status(400);
    res.json({
      error: `Child path not in our root, [${localImagesRootPath}].`,
    });
    return;
  }

  console.log("uploading all images from folder", iPath);

  const db = useDatabase();

  const newImportId = await db.addLocalImageImportPath(iPath, iCategory, iAuthor, iAuthorUrl, tags);
  if (newImportId) {
    console.log("IMPORTING FILES:", newImportId);
  } else {
    console.log("Skipping import, folder already exists for category");
  }

  //TODO: only do below if newImportId
  const files = await glob(join(iPath, "**/*"));
  const filenameList = files
    .filter((fn) => supportedLocalImageExtensions.includes(fn.toLowerCase().split(".").pop() || ""))
    .map((fn) => relative(localImagesRootPath, fn));
  const images = await db.addLocalImages(filenameList, iAuthor, iAuthorUrl);

  if (images !== undefined) {
    await db.addImagesToCategory(iCategory, images, tags);
  } else {
    res.status(400);
    res.json({
      error: `No image IDs were returned.`,
    });
    return;
  }

  res.json({
    ok: true,
  });
});

router.get("/sources", async (req: Request, res: Response) => {
  const db = useDatabase();

  const sources = await db.getImageSources();

  res.send(sources);
});

router.delete("/unused", needAdmin, async (req: Request, res: Response) => {
  const db = useDatabase();

  await db.deleteUnusedImages();

  res.json({
    ok: true,
  });
});
