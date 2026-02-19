import path from "path";
import { exec } from "child_process";
import fs from "fs";
import mime from "mime-types";
import { S3Client, PutObjectCommand } from "@aws-sdk/client-s3";

const PROJECT_ID = process.env.PROJECT_ID;

// s3 related credentials
const S3client = new S3Client({
  region: process.env.REGION,
  endpoint: process.env.END_POINT,
  credentials: {
    accessKeyId: process.env.accessKeyId,
    secretAccessKey: process.env.secretAccessKey,
  },
});

async function main() {
  console.log("executing script.js file");
  const outerDirPath = path.join(__dirname, "output");
  try {
    const cmd = exec(
      `cd ${outerDirPath} && npm install -g pnpm && pnpm install && pnpm run build`,
    );
    cmd.stdout.on("data", (data) => {
      console.log("data", data.toString());
    });
    cmd.stderr.on("error", (error) => {
      console.log("error", error.toString());
    });
    cmd.on("close", async function () {
      console.log("build complete");
      const distOutputFolder = path.join(__dirname, "output", "dist");
      /*we need to read the files synchronously, to make sure every file is there
      going to give me return of array */
      const distContent = fs.readdirSync(distOutputFolder, { recursive: true });
      for (const filepath of distContent) {
        if (fs.lstatSync(filepath).isDirectory()) continue;
        console.log("uploading files into s3 buckets");
        const input = {
          Bucket: BUCKET_NAME,
          //its means how you going to store things into s3-(folderName),
          Key: `__outputs/${PROJECT_ID}/${filepath}`,
          // what going to be store into keys or folder
          Body: fs.createReadStream(filepath),
          // dynamic telling the s3 that content can be anything with using mime,
          ContentType: mime.lookup(`${filepath}`),
        };
        const command = new PutObjectCommand(input);
        const response = await S3client.send(command);
        console.log("file are succesfully uploaded into s3");
        console.log(response);
      }
    });
  } catch (error) {
    console.log(error);
  }
}
