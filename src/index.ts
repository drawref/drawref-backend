import express, { Express } from "express";
import cors from "cors";

import { initialiseDatabase } from "./db/database.js";
import { authenticate } from "./auth/bearerMiddleware.js";
import { port, databaseUrl } from "./config/env.js";
import { router as authRouter } from "./routes/auth.js";
import { router as categoriesRouter } from "./routes/categories.js";
import { router as imagesRouter } from "./routes/images.js";
import { router as samplesRouter } from "./routes/samples.js";
import { router as sessionRouter } from "./routes/session.js";
import { router as rootRouter } from "./routes/root.js";
import { router as userRouter } from "./routes/user.js";
import { router as uploadRouter } from "./routes/upload.js";
import { confirmS3Works } from "./files/s3.js";

if (!(await confirmS3Works())) {
  console.log("Could not connect to S3 and verify bucket connection.");
  console.log("Closing server");
  process.exit();
}

initialiseDatabase(databaseUrl);

const app: Express = express();

app.use(authenticate);
app.use(cors());
app.use(express.json());

app.use("/upload", uploadRouter);

app.use("/api/auth", authRouter);
app.use("/api/categories", categoriesRouter);
app.use("/api/image", imagesRouter);
app.use("/api/samples", samplesRouter);
app.use("/api/session", sessionRouter);
app.use("/api/user", userRouter);
app.use("/api", rootRouter);

// intended mostly for dev testing
app.use("/images", express.static("public"));

app.listen(port, () => {
  console.log(`[server]: Server is running at http://localhost:${port}`);
});
