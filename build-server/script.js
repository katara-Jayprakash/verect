/**
 --> for build server:
    1. first go to folder
    2. run the command like npm install -g pnpm and pnpm install and  pnpm run build
    3.  upload the dist folder (build output ) to s3 bucket
 -->for logs part
     1. use publisher and subsriber architecture of redis and put all the build logs into redis
 */

import path from "path";
import { fileURLToPath } from "url";
import { dirname } from "path";
import fs from "fs";
import mime from "mime-types";
import { exec } from "child_process";
import { Upload } from "@aws-sdk/lib-storage";
import { S3Client, S3 } from "@aws-sdk/client-s3";
import dotenv from "dotenv";
import redis from "ioredis";
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

// Initialize Redis publisher
const redisPublisher = new redis(process.env.serviceUri);

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

async function publishLogs(message) {
  const logEntry = {
    timestamp: new Date().toISOString(),
    message: message,
    projectId: PROJECT_ID,
  };
  try {
    await redisPublisher.publish(
      `build-logs:${PROJECT_ID}`,
      JSON.stringify(logEntry),
    );
  } catch (error) {
    console.error("Redis publish error:", error.message);
  }
}

async function main() {
  await publishLogs(`Building project: ${PROJECT_ID}`);
  await publishLogs(`Building project: ${PROJECT_ID}`);
  console.log(" exectuing script.js file ");
  const outputPath = path.join(__dirname, "output");
  try {
    await publishLogs("Installing dependencies and building project...");
    const cmd = exec(
      `cd ${outputPath} && npm install -g pnpm && pnpm install && pnpm run build`,
    );
    cmd.stdout.on("data", async (data) => {
      console.log("data", data.toString());
      await publishLogs(data.toString());
    });

    cmd.stderr.on("data", async (data) => {
      console.log("error", data.toString());
      await publishLogs(`Error: ${data.toString()}`);
      await redisPublisher.quit();
    });

    cmd.on("close", async function (code) {
      if (code !== 0) {
        console.error(`Build failed with exit code ${code}`);
        await publishLogs(`Build failed with exit code ${code}`);
        return;
      }
      console.log("build completed and we have successfully got dist folder");
      await publishLogs("Build completed");
      const distOutputFolder = path.join(__dirname, "output", "dist");
      // reading directory synchronous using readdirSync which return array of files
      const distContent = fs.readdirSync(distOutputFolder, { recursive: true });
      await publishLogs("starting uploading ");
      for (const filepath of distContent) {
        const fullPath = path.join(distOutputFolder, filepath);
        if (fs.lstatSync(fullPath).isDirectory()) continue;
        console.log(`Uploading: ${filepath}`);
        try {
          const s3key = `__outputs/${PROJECT_ID}/${filepath}`;
          await publishLogs(`Uploading: ${filepath}`);
          await uploadToS3(fullPath, s3key);
          console.log("succesfully uploded things to the s3 bucket", filepath);
          await publishLogs(`Uploaded: ${filepath}`);
        } catch (error) {
          console.error("S3 Operation Failed");
          console.error("Message:", error.message);
          console.error("Name:", error.name);
          console.error(`something fucked up ${error}`);
        }
      }
      await publishLogs("Build process finished");
      await redisPublisher.quit();
    });
  } catch (error) {
    console.error(`here is the error`, error);
    await redisPublisher.quit();
  }
}
main()
  .then(() => {
    console.log("things are uploaded into s3 finally! ");
  })
  .catch(async (err) => {
    console.error("Script failed:", err);
    await publishLogs(`Fatal error: ${err.message}`);
    await redisPublisher.quit();
  });
