const fs = require("node:fs");
const readline = require("node:readline");
const { PlayBeatmap } = require("@tosuapp/lazer-calculator-prebuilt");

const variants = new Map();
const cache = new Map();
let nextId = 1;

function normaliseMods(mods, isLazer) {
  const result = (mods ?? [])
    .filter((mod) => mod.acronym !== "LZ")
    .map((mod) => ({
      acronym: mod.acronym,
      settings: new Map(Object.entries(mod.settings ?? {})),
    }));

  if (!isLazer && !result.some((mod) => mod.acronym === "CL")) {
    result.push({ acronym: "CL", settings: new Map() });
  }

  return result;
}

function loadBeatmap(request) {
  const cacheKey = JSON.stringify([
    request.path,
    request.mods ?? [],
    request.isLazer,
  ]);
  const cachedId = cache.get(cacheKey);

  if (cachedId !== undefined) {
    const cached = variants.get(cachedId);
    return {
      id: cachedId,
      attributes: cached.attributeData,
      strains: cached.strains,
    };
  }

  const beatmap = PlayBeatmap.parse(fs.readFileSync(request.path, "utf8"));
  beatmap.applyMods(normaliseMods(request.mods, request.isLazer));

  const gradual = beatmap.createGradualDifficulty();
  const attributes = [];
  const attributeData = [];

  while (gradual.advance()) {
    const current = gradual.createDifficultyAttrs();
    attributes.push(current);
    attributeData.push(current.getData());
  }

  const currentStrains = gradual.getCurrentStrains();
  const strains = {
    aim: Array.from(currentStrains.aim.value),
    speed: Array.from(currentStrains.speed.value),
    flashlight: Array.from(currentStrains.flashlight.value),
    reading: Array.from(currentStrains.reading.value),
    total: Array.from(currentStrains.strains.value),
  };

  const id = nextId++;
  variants.set(id, { beatmap, attributes, attributeData, strains });
  cache.set(cacheKey, id);

  return { id, attributes: attributeData, strains };
}

function calculatePerformance(request) {
  const variant = variants.get(request.id);

  if (!variant) throw new Error(`Unknown beatmap variant ${request.id}`);

  const index = Math.max(
    0,
    Math.min(request.index, variant.attributes.length - 1),
  );
  const score = {
    totalScore: 0,
    isLegacyScore: false,
    accuracy: 0,
    maxCombo: 0,
    sliderEndHits: 0,
    comboBreaks: 0,
    ignoreHits: 0,
    ignoreMisses: 0,
    largeBonuses: 0,
    smallBonuses: 0,
    largeTickHits: 0,
    largeTickMisses: 0,
    smallTickHits: 0,
    smallTickMisses: 0,
    perfects: 0,
    greats: 0,
    goods: 0,
    oks: 0,
    mehs: 0,
    misses: 0,
    ...request.score,
  };

  score.accuracy = variant.beatmap.calculateAccuracy(score);

  return {
    accuracy: score.accuracy,
    performance: variant.beatmap.calculatePerformance(
      variant.attributes[index],
      score,
    ),
  };
}

function handle(request) {
  switch (request.command) {
    case "load":
      return loadBeatmap(request);
    case "performance":
      return calculatePerformance(request);
    case "ping":
      return { version: "20260729" };
    case "clear":
      variants.clear();
      cache.clear();
      return {};
    default:
      throw new Error(`Unknown command ${request.command}`);
  }
}

const input = readline.createInterface({ input: process.stdin });

input.on("line", (line) => {
  try {
    const request = JSON.parse(line);
    process.stdout.write(`${JSON.stringify({ ok: true, result: handle(request) })}\n`);
  } catch (error) {
    process.stdout.write(
      `${JSON.stringify({ ok: false, error: error?.stack ?? String(error) })}\n`,
    );
  }
});
