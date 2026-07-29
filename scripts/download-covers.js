/**
 * Downloads cover images from Pexels.
 *
 * Usage:
 *   1. Register a free API key at https://www.pexels.com/api/
 *   2. Set the environment variable: set PEXELS_API_KEY=your_key   (Windows cmd)
 *                                    $env:PEXELS_API_KEY="your_key" (PowerShell)
 *                                    export PEXELS_API_KEY=your_key  (Git Bash / Linux / macOS)
 *   3. Run: node scripts/download-covers.js
 *
 * Optional arguments:
 *   --count=N     Total number of images to download (default 9)
 *   --dir=PATH    Output directory (default app/appearance/covers)
 */

const fs = require("fs");
const path = require("path");
const sharp = require(path.resolve(__dirname, "..", "app", "node_modules", "sharp"));

const API_KEY = process.env.PEXELS_API_KEY;
if (!API_KEY) {
    console.error("Please set the PEXELS_API_KEY environment variable.");
    console.error("Free registration: https://www.pexels.com/api/");
    process.exit(1);
}

// Parse command-line arguments.
const args = process.argv.slice(2);
const getArg = (name, fallback) => {
    const found = args.find(a => a.startsWith(`--${name}=`));
    return found ? found.split("=")[1] : fallback;
};
const TOTAL = parseInt(getArg("count", "72"), 10);
const OUT_DIR = path.resolve(__dirname, "..", getArg("dir", "app/appearance/covers"));

// Search categories: take an equal number of images from each category.
const SEARCH_QUERIES = [
    { query: "epic mountain landscape photography", label: "Nature", key: "coverNature" },
    { query: "city night skyline blue hour", label: "City night", key: "coverCityNight" },
    { query: "classical architecture cathedral historic", label: "Classical architecture", key: "coverClassicalArchitecture" },
    { query: "cozy reading nook books candle", label: "Reading nook", key: "coverReadingNook" },
    { query: "zen garden minimal calm aesthetic", label: "Zen minimalism", key: "coverZenMinimal" },
    { query: "architecture light shadow geometry", label: "Light geometry", key: "coverLightGeometry" },
    { query: "winding road path journey landscape", label: "Road ahead", key: "coverRoadAhead" },
    { query: "autumn fall leaves colorful forest", label: "Autumn leaves", key: "coverAutumnLeaves" },
    { query: "neon lights night city vibrant colorful", label: "Neon nights", key: "coverNeonNights" },
    { query: "desert sand dune arid landscape", label: "Desert", key: "coverDesert" },
    { query: "aurora borealis northern lights sky", label: "Aurora", key: "coverAurora" },
    { query: "morning mist fog valley mountain", label: "Misty morning", key: "coverMistyMorning" },
    { query: "countryside rural farm meadow peaceful", label: "Countryside", key: "coverCountryside" },
    { query: "tea ceremony calligraphy writing desk", label: "Tea ceremony", key: "coverTeaCeremony" },
    { query: "calm lake reflection mirror water still", label: "Still water", key: "coverStillWater" },
    { query: "chinese garden pavilion architecture", label: "Chinese garden", key: "coverChineseGarden" },
    { query: "karst mountain mist landscape china", label: "Ink-wash landscape", key: "coverInkWashLandscape" },
    { query: "wildlife animal deer fox bird nature", label: "Wildlife", key: "coverWildlife" },
];
const PER_QUERY = Math.ceil(TOTAL / SEARCH_QUERIES.length);

/**
 * Searches images through the Pexels API.
 */
async function searchPhotos(query, perPage) {
    const url = `https://api.pexels.com/v1/search?query=${encodeURIComponent(query)}&per_page=${perPage}&orientation=landscape&size=medium`;
    const resp = await fetch(url, {
        headers: { Authorization: API_KEY },
    });
    if (!resp.ok) {
        const text = await resp.text();
        throw new Error(`Pexels API request failed (${resp.status}): ${text}`);
    }
    const data = await resp.json();
    return data.photos || [];
}

/**
 * Downloads a single image.
 */
async function downloadPhoto(photo, outDir, index) {
    // Fetch the original from Pexels and crop it to a 2x Retina size with sharp.
    const width = 2400;
    const height = 800;
    const imgUrl = photo.src.original;

    const filename = `cover_${String(index).padStart(3, "0")}.webp`;
    const filePath = path.join(outDir, filename);

    console.log(`  Downloading: ${filename} <- ${photo.photographer} / Pexels`);

    const imgResp = await fetch(imgUrl);
    if (!imgResp.ok) {
        throw new Error(`Image download failed (${imgResp.status}): ${imgUrl}`);
    }
    const inputBuffer = Buffer.from(await imgResp.arrayBuffer());

    // Convert to webp.
    const outputBuffer = await sharp(inputBuffer)
        .resize(width, height, { fit: "cover" })
        .webp({ quality: 85 })
        .toBuffer();

    fs.writeFileSync(filePath, outputBuffer);

    const kb = (outputBuffer.length / 1024).toFixed(1);
    console.log(`     ${filename} (${kb} KB)`);

    return {
        file: filename,
        category: photo.alt || "",
        photographer: photo.photographer,
        photographer_url: photo.photographer_url,
        pexels_url: photo.url,
        width,
        height,
    };
}

async function main() {
    console.log(`Starting cover image download (${TOTAL} total)...\n`);

    // Create the output directory.
    fs.mkdirSync(OUT_DIR, { recursive: true });

    // Global dedupe set by photo id.
    const seen = new Set();
    let allPhotos = [];

    for (const { query, label, key } of SEARCH_QUERIES) {
        console.log(`Searching "${label}"...`);
        const photos = await searchPhotos(query, PER_QUERY * 2); // Fetch extras so enough remain after dedupe.
        // Dedupe by global ID and per-category photographer.
        let categoryPhotos = [];
        const categoryPhotographers = new Set();
        for (const p of photos) {
            if (seen.has(p.id)) continue;
            if (categoryPhotographers.has(p.photographer)) continue; // Avoid repeating photographers in one category.
            seen.add(p.id);
            categoryPhotographers.add(p.photographer);
            p._category = key;
            categoryPhotos.push(p);
        }
        // Take PER_QUERY images from each category.
        categoryPhotos = categoryPhotos.slice(0, PER_QUERY);
        allPhotos = allPhotos.concat(categoryPhotos);
        console.log(`   Found ${photos.length}; selected ${categoryPhotos.length} after dedupe`);
    }

    if (allPhotos.length < TOTAL) {
        console.warn(`Only found ${allPhotos.length} images (target ${TOTAL}); downloading all available images.`);
    }

    // Download images.
    console.log(`\nDownloading ${allPhotos.length} images to ${OUT_DIR} ...\n`);
    const manifest = [];
    for (let i = 0; i < allPhotos.length; i++) {
        try {
            const entry = await downloadPhoto(allPhotos[i], OUT_DIR, i + 1);
            entry.category = allPhotos[i]._category;
            manifest.push(entry);
        } catch (err) {
            console.error(`   Download failed: ${err.message}`);
        }
    }

    // Write manifest.
    const manifestPath = path.join(OUT_DIR, "manifest.json");
    fs.writeFileSync(manifestPath, JSON.stringify(manifest, null, "  "), "utf-8");
    console.log(`\nmanifest.json generated (${manifest.length} entries)`);

    console.log("\nDownload complete. Images are located at:");
    console.log(`   ${OUT_DIR}`);
    console.log("\nOpen the images directly in a browser to review them.");
    console.log("After approval, update Background.ts to integrate them into the cover dialog.\n");
}

main().catch(err => {
    console.error("Script failed:", err.message);
    process.exit(1);
});
