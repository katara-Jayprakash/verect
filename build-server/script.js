import { path } from "path";
import { exec } from "child_process";

async function main() {
  console.log("executing script.js files");
  const outputPath = path.join(__dirname, "output");
  const cmd = exec(
    `cd ${outputPath} && npm install -g pnpm && pnpm install && pnpm run build`,
  );
  cmd.stdout.on("data", function (data) {
    console.log(data.toString());
  });
  cmd.stdout.on("e", function (e) {
    console.log(data.toString());
  });
  cmd.on("close", function () {
    console.log("build complete");
  });
}
