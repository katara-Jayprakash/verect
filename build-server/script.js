//steps i have to implement
// 1 . cd into the output folder,
// 2. install the pnpm and globally and run pnpm install and then run pnpm build command
// 3. take all the build dict to the s3 bucket
import path from "path";
import { exec } from "child_process";
import fs from "fs";
import { S3Client, GetObjectCommand } from "@aws-sdk/client-s3";

async function main() {
  console.log("executing script.js file");
  const outerDirPath = path.join(__dirname, "output");
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
    // we need to read the files synchronously, to make sure every file is there
    // going to give me return of array
    const distContent = fs.readdirSync(distOutputFolder, { recursive: true });
    for (const filepath of distContent) {
      // removing folder from filepath, we dont need then to push it into our s3
      if (fs.lstatSync(filepath).isDirectory()) continue;
    }
  });
}
