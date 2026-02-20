//first go to folder
// run the command like npm install -g pnpm and pnpm install and then pnpm run build
// after then but now we have to do one more thing first pick all the files and make them into recrusive send all the build to s3 bucket
import path from "path";
import { fileURLToPath } from "url";
import { dirname } from "path";
import fs from "fs";
import mime from "mime-types";
import { exec } from "child_process";
import { Upload } from "@aws-sdk/lib-storage";
import { S3Client, S3 } from "@aws-sdk/client-s3";
import dotenv from "dotenv";
dotenv.config();

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);
const PROJECT_ID = process.env.PROJECT_ID;

const S3client = new S3Client({
  region: process.env.REGION,
  endpoint: process.env.END_POINT,
  credentials: {
    accessKeyId: process.env.accessKeyId,
    secretAccessKey: process.env.secretAccessKey,
  },
  forcePathStyle: true,
});

async function uploadToS3(fullPath, s3key) {
  const fileStream = fs.createReadStream(fullPath);

  const parallelUploads = new Upload({
    client: S3client,
    params: {
      Bucket: process.env.BUCKET_NAME,
      Key: s3key,
      Body: fileStream,
      ContentType: mime.lookup(fullPath) || "application/octet-stream",
    },
    queueSize: 5,
    partSize: 1024 * 1024 * 8,
    leavePartsOnError: false,
  });

  parallelUploads.on("httpUploadProgress", (progress) => {
    console.log(`Uploaded ${progress.loaded} bytes for ${s3key}`);
  });
  return parallelUploads.done();
}

async function main() {
  console.log(" exectuing script.js file ");
  const outputPath = path.join(__dirname, "output");
  try {
    const cmd = exec(
      `cd ${outputPath} && npm install -g pnpm && pnpm install && pnpm run build`,
    );
    cmd.stdout.on("data", (data) => {
      console.log("data", data.toString());
    });

    cmd.stderr.on("data", (data) => {
      console.log("error", data.toString());
    });

    cmd.on("close", async function (code) {
      if (code !== 0) {
        console.error(`Build failed with exit code ${code}`);
        return;
      }
      console.log("build completed and we have successfully got dist folder");
      const distOutputFolder = path.join(__dirname, "output", "dist");
      // reading directory synchronous using readdirSync which return array of files
      const distContent = fs.readdirSync(distOutputFolder, { recursive: true });
      for (const filepath of distContent) {
        const fullPath = path.join(distOutputFolder, filepath);
        if (fs.lstatSync(fullPath).isDirectory()) continue;
        console.log(`Uploading: ${filepath}`);
        try {
          const s3key = `__outputs/${PROJECT_ID}/${filepath}`;
          await uploadToS3(fullPath, s3key);
          console.log("succesfully uploded things to the s3 bucket", filepath);
        } catch (error) {
          console.error("S3 Operation Failed");
          console.error("Message:", error.message);
          console.error("Name:", error.name);
          console.error(`something fucked up ${error}`);
        }
      }
    });
  } catch (error) {
    console.error(`here is the error`, error);
  }
}
main()
  .then(() => {
    console.log("things are uploaded into s3 finally! ");
  })
  .catch((err) => {
    console.error("Script failed:", err);
  });
