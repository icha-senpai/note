const fs = require("fs");
const path = require("path");
const sharp = require("../app/node_modules/sharp");

const root = path.resolve(__dirname, "..");
const sourceLogo = path.join(root, "scribli.png");
const iconCrop = {left: 300, top: 300, width: 340, height: 340};

const write = async (relativePath, buffer) => {
  const filePath = path.join(root, relativePath);
  await fs.promises.mkdir(path.dirname(filePath), {recursive: true});
  await fs.promises.writeFile(filePath, buffer);
};

const renderIconPng = async (width, height = width) =>
  sharp(sourceLogo)
    .extract(iconCrop)
    .resize(width, height, {fit: "contain"})
    .png({compressionLevel: 9})
    .toBuffer();

const renderLogoPng = async (width, height) =>
  sharp(sourceLogo)
    .resize(width, height, {fit: "contain", background: {r: 0, g: 0, b: 0, alpha: 0}})
    .png({compressionLevel: 9})
    .toBuffer();

const pngDataSvg = (png, width, height) => Buffer.from(`<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="${width}" height="${height}" viewBox="0 0 ${width} ${height}">
  <image href="data:image/png;base64,${png.toString("base64")}" width="${width}" height="${height}"/>
</svg>
`);

const makeLoadingSvg = (iconPng) => Buffer.from(`<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" width="64" height="64" viewBox="0 0 64 64">
  <defs>
    <linearGradient id="spin" x1="8" y1="8" x2="56" y2="56">
      <stop offset="0" stop-color="#00d8ff"/>
      <stop offset="0.55" stop-color="#0b6dff"/>
      <stop offset="1" stop-color="#a34cff"/>
    </linearGradient>
  </defs>
  <circle cx="32" cy="32" r="27" fill="none" stroke="#071225" stroke-width="4" opacity="0.7"/>
  <path d="M32 5a27 27 0 0 1 25 17" fill="none" stroke="url(#spin)" stroke-width="5" stroke-linecap="round">
    <animateTransform attributeName="transform" type="rotate" calcMode="linear" values="0 32 32;360 32 32" keyTimes="0;1" dur="1.2s" repeatCount="indefinite"/>
  </path>
  <image href="data:image/png;base64,${iconPng.toString("base64")}" x="16" y="16" width="32" height="32"/>
</svg>
`);

const makeIco = async () => {
  const sizes = [16, 24, 32, 48, 64, 128, 256];
  const images = [];
  for (const size of sizes) {
    images.push({size, buffer: await renderIconPng(size)});
  }

  const headerSize = 6 + images.length * 16;
  let offset = headerSize;
  const header = Buffer.alloc(headerSize);
  header.writeUInt16LE(0, 0);
  header.writeUInt16LE(1, 2);
  header.writeUInt16LE(images.length, 4);

  images.forEach((image, index) => {
    const entry = 6 + index * 16;
    header.writeUInt8(image.size === 256 ? 0 : image.size, entry);
    header.writeUInt8(image.size === 256 ? 0 : image.size, entry + 1);
    header.writeUInt8(0, entry + 2);
    header.writeUInt8(0, entry + 3);
    header.writeUInt16LE(1, entry + 4);
    header.writeUInt16LE(32, entry + 6);
    header.writeUInt32LE(image.buffer.length, entry + 8);
    header.writeUInt32LE(offset, entry + 12);
    offset += image.buffer.length;
  });

  return Buffer.concat([header, ...images.map((image) => image.buffer)]);
};

const makeIcns = async () => {
  const entries = [
    ["ic07", 128],
    ["ic08", 256],
    ["ic09", 512],
    ["ic10", 1024],
  ];
  const chunks = [];
  for (const [type, size] of entries) {
    const png = await renderIconPng(size);
    const chunkHeader = Buffer.alloc(8);
    chunkHeader.write(type, 0, 4, "ascii");
    chunkHeader.writeUInt32BE(png.length + 8, 4);
    chunks.push(chunkHeader, png);
  }
  const totalLength = 8 + chunks.reduce((sum, chunk) => sum + chunk.length, 0);
  const header = Buffer.alloc(8);
  header.write("icns", 0, 4, "ascii");
  header.writeUInt32BE(totalLength, 4);
  return Buffer.concat([header, ...chunks]);
};

const makeBmp24 = async (png, width, height) => {
  const raw = await sharp(png)
    .resize(width, height, {fit: "cover"})
    .flatten({background: "#11161c"})
    .raw()
    .toBuffer();
  const rowSize = Math.ceil((width * 3) / 4) * 4;
  const pixelSize = rowSize * height;
  const header = Buffer.alloc(54);
  header.write("BM", 0, 2, "ascii");
  header.writeUInt32LE(54 + pixelSize, 2);
  header.writeUInt32LE(54, 10);
  header.writeUInt32LE(40, 14);
  header.writeInt32LE(width, 18);
  header.writeInt32LE(height, 22);
  header.writeUInt16LE(1, 26);
  header.writeUInt16LE(24, 28);
  header.writeUInt32LE(pixelSize, 34);

  const pixels = Buffer.alloc(pixelSize);
  for (let y = 0; y < height; y++) {
    const sourceY = height - 1 - y;
    for (let x = 0; x < width; x++) {
      const source = (sourceY * width + x) * 3;
      const target = y * rowSize + x * 3;
      pixels[target] = raw[source + 2];
      pixels[target + 1] = raw[source + 1];
      pixels[target + 2] = raw[source];
    }
  }
  return Buffer.concat([header, pixels]);
};

const findFiles = async (directory, matcher, results = []) => {
  const entries = await fs.promises.readdir(directory, {withFileTypes: true});
  for (const entry of entries) {
    const fullPath = path.join(directory, entry.name);
    if (entry.isDirectory()) {
      await findFiles(fullPath, matcher, results);
    } else if (matcher(fullPath)) {
      results.push(fullPath);
    }
  }
  return results;
};

const appxTargets = [
  ["app/appx/assets/Square150x150Logo.png", 150, 150, "icon"],
  ["app/appx/assets/Square150x150Logo.targetsize-150_altform-lightunplated.png", 150, 150, "icon"],
  ["app/appx/assets/Square150x150Logo.targetsize-150_altform-unplated.png", 150, 150, "icon"],
  ["app/appx/assets/Square44x44Logo.png", 44, 44, "icon"],
  ["app/appx/assets/Square44x44Logo.targetsize-16_altform-lightunplated.png", 16, 16, "icon"],
  ["app/appx/assets/Square44x44Logo.targetsize-16_altform-unplated.png", 16, 16, "icon"],
  ["app/appx/assets/Square44x44Logo.targetsize-20_altform-lightunplated.png", 20, 20, "icon"],
  ["app/appx/assets/Square44x44Logo.targetsize-20_altform-unplated.png", 20, 20, "icon"],
  ["app/appx/assets/Square44x44Logo.targetsize-24_altform-lightunplated.png", 24, 24, "icon"],
  ["app/appx/assets/Square44x44Logo.targetsize-24_altform-unplated.png", 24, 24, "icon"],
  ["app/appx/assets/Square44x44Logo.targetsize-256_altform-lightunplated.png", 256, 256, "icon"],
  ["app/appx/assets/Square44x44Logo.targetsize-256_altform-unplated.png", 256, 256, "icon"],
  ["app/appx/assets/Square44x44Logo.targetsize-30_altform-lightunplated.png", 30, 30, "icon"],
  ["app/appx/assets/Square44x44Logo.targetsize-30_altform-unplated.png", 30, 30, "icon"],
  ["app/appx/assets/Square44x44Logo.targetsize-32_altform-lightunplated.png", 32, 32, "icon"],
  ["app/appx/assets/Square44x44Logo.targetsize-32_altform-unplated.png", 32, 32, "icon"],
  ["app/appx/assets/Square44x44Logo.targetsize-36_altform-lightunplated.png", 36, 36, "icon"],
  ["app/appx/assets/Square44x44Logo.targetsize-36_altform-unplated.png", 36, 36, "icon"],
  ["app/appx/assets/Square44x44Logo.targetsize-40_altform-lightunplated.png", 40, 40, "icon"],
  ["app/appx/assets/Square44x44Logo.targetsize-40_altform-unplated.png", 40, 40, "icon"],
  ["app/appx/assets/Square44x44Logo.targetsize-48_altform-lightunplated.png", 48, 48, "icon"],
  ["app/appx/assets/Square44x44Logo.targetsize-48_altform-unplated.png", 48, 48, "icon"],
  ["app/appx/assets/Square44x44Logo.targetsize-60_altform-lightunplated.png", 60, 60, "icon"],
  ["app/appx/assets/Square44x44Logo.targetsize-60_altform-unplated.png", 60, 60, "icon"],
  ["app/appx/assets/Square44x44Logo.targetsize-64_altform-lightunplated.png", 64, 64, "icon"],
  ["app/appx/assets/Square44x44Logo.targetsize-64_altform-unplated.png", 64, 64, "icon"],
  ["app/appx/assets/Square44x44Logo.targetsize-72_altform-lightunplated.png", 72, 72, "icon"],
  ["app/appx/assets/Square44x44Logo.targetsize-72_altform-unplated.png", 72, 72, "icon"],
  ["app/appx/assets/Square44x44Logo.targetsize-80_altform-lightunplated.png", 80, 80, "icon"],
  ["app/appx/assets/Square44x44Logo.targetsize-80_altform-unplated.png", 80, 80, "icon"],
  ["app/appx/assets/Square44x44Logo.targetsize-96_altform-lightunplated.png", 96, 96, "icon"],
  ["app/appx/assets/Square44x44Logo.targetsize-96_altform-unplated.png", 96, 96, "icon"],
  ["app/appx/assets/StoreLogo.png", 50, 50, "icon"],
  ["app/appx/assets/StoreLogo.targetsize-256_altform-lightunplated.png", 256, 256, "icon"],
  ["app/appx/assets/StoreLogo.targetsize-256_altform-unplated.png", 256, 256, "icon"],
  ["app/appx/assets/Wide310x150Logo.png", 310, 150, "logo"],
];

(async () => {
  if (!fs.existsSync(sourceLogo)) {
    throw new Error(`Missing source logo: ${sourceLogo}`);
  }

  const icon512 = await renderIconPng(512);
  const icon256 = await renderIconPng(256);
  const icon160 = await renderIconPng(160);
  const logo1536 = await renderLogoPng(1536, 1024);
  const loadingIcon = await renderIconPng(128);

  await write("app/src/assets/icon.png", icon512);
  await write("app/src/assets/icon256.png", icon256);
  await write("app/src/assets/icon-mac.png", icon512);
  await write("app/electron/icon.png", icon512);
  await write("app/stage/icon.png", icon512);
  await write("app/stage/icon-large.png", icon512);
  await write("app/stage/images/icon.png", icon160);

  await write("app/src/assets/logo.png", logo1536);
  await write("app/stage/logo.png", logo1536);
  await write("app/src/assets/icon.svg", pngDataSvg(icon512, 512, 512));
  await write("app/stage/icon.svg", pngDataSvg(icon512, 512, 512));
  await write("app/src/assets/logo.svg", pngDataSvg(logo1536, 1536, 1024));
  await write("app/stage/logo.svg", pngDataSvg(logo1536, 1536, 1024));
  await write("app/stage/loading-pure.svg", makeLoadingSvg(loadingIcon));
  await write("app/stage/loading.svg", makeLoadingSvg(loadingIcon));

  for (const size of [16, 32, 48, 64, 128, 256, 512]) {
    await write(`app/src/assets/icon/${size}x${size}.png`, await renderIconPng(size));
  }

  for (const [target, width, height, source] of appxTargets) {
    await write(target, source === "logo" ? await renderLogoPng(width, height) : await renderIconPng(width, height));
  }

  await write("app/src/assets/icon.ico", await makeIco());
  await write("app/src/assets/icon.icns", await makeIcns());
  await write("app/nsis/installerSidebar.bmp", await makeBmp24(await renderLogoPng(164, 314), 164, 314));
  await write("app/nsis/uninstallerSidebar.bmp", await makeBmp24(await renderLogoPng(164, 314), 164, 314));

  const guideLogoFiles = await findFiles(path.join(root, "app", "guide"), (filePath) =>
    /[\\/]assets[\\/]siyuan-128-.*\.png$/i.test(filePath)
  );
  for (const filePath of guideLogoFiles) {
    await fs.promises.writeFile(filePath, await renderIconPng(128));
  }
})();
