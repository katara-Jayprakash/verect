//steps i have to implement
// 1 . cd into the output folder,
// 2. install the pnpm and globally and run pnpm install and then run pnpm build command
// 3. take all the build dict to the s3 bucket
import path from "path";
import { exec } from "child_process";
import fs from "fs";
import { S3Client, PutObjectCommand, Bucket$ } from "@aws-sdk/client-s3";
import BodyReadable from "undici-types/readable";
import { MIMEType } from "util";

const PROJECT_ID = process.env.PROJECT_ID;
const BUCKET_NAME = process.env.BUCKET_NAME;
const END_POINT = process.env.Endpoint;

const S3client = new S3Client({
  region: END_POINT,
  credentials: {
    accessKeyId: "",
    secretAccessKey: "",
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
        if (fs.lstatSync(filepath).isDirectory()) continue; // removing folder from filepath,
        // putting things into s3 bucket
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
        console.log(response);
      }
    });
  } catch (error) {
    console.log(error);
  }
}
