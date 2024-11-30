import dotenv from "dotenv";
import { V3 } from "paseto";

dotenv.config();

// serving ports and urls
export const port = process.env.PORT || 3300;
export const myURL = process.env.MY_BASEURL || "";
export const frontendURL = process.env.FRONTEND_BASEURL || "";

// db
if (!process.env.DATABASE_URL) {
  console.log("Postgres database must be defined in the DATABASE_URL environment variable.");
  console.log("Closing server");
  process.exit();
}
export const databaseUrl = process.env.DATABASE_URL;

// image uploading
export const uploadPathTmp = process.env.TMP_UPLOAD_PATH || "./tmp";
export const uploadBucket = process.env.UPLOAD_S3_BUCKET;
export const uploadKeyPrefix = process.env.UPLOAD_KEY_PREFIX || "/";
export const uploadUrlPrefix = process.env.UPLOAD_URL_PREFIX || "";

if (!uploadBucket || uploadUrlPrefix === "") {
  console.log("Upload details must be defined in the UPLOAD_S3_BUCKET and UPLOAD_URL_PREFIX environment variables.");
  console.log("Closing server");
  process.exit();
}

// auth
export const githubClientId = process.env.GITHUB_CLIENT_ID;
export const githubClientSecret = process.env.GITHUB_CLIENT_SECRET;
export const githubAdminUids = (process.env.GITHUB_ADMIN_UIDS || "").split(" ").map((uid) => parseInt(uid));
export const githubActive = githubClientId && githubClientSecret;

// crypto
if (!process.env.PASETO_LOCAL_KEY) {
  console.log("PASETO local key must be defined in the PASETO_LOCAL_KEY environment variable.");
  console.log("Here's a new randomly-generated one you can use:");
  console.log(
    await V3.generateKey("local", {
      format: "paserk",
    }),
  );
  console.log("Closing server");
  process.exit();
}
export const pasetoLocalKey = process.env.PASETO_LOCAL_KEY;
