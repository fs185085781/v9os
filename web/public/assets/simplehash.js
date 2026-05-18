const simpleFileHash = async function (file, options = {}) {
  const simpleHash = this;
  if (!simpleHash.smartFileSampler) {
    simpleHash.smartFileSampler = async function (
      file,
      numberOfSamples = 200,
      sampleSize = 1024,
    ) {
      const fileSize = file.size;
      const samples = [];
      if (fileSize <= numberOfSamples * sampleSize) {
        const entireFileBuffer = await file.arrayBuffer();
        return [entireFileBuffer];
      }
      for (let i = 0; i < numberOfSamples; i++) {
        const startPos = Math.floor((i * fileSize) / numberOfSamples);
        const endPos = Math.min(startPos + sampleSize, fileSize);
        const chunk = file.slice(startPos, endPos);
        const chunkBuffer = await chunk.arrayBuffer();
        samples.push(chunkBuffer);
      }
      return samples;
    };
  }
  if (!simpleHash.fallbackSimpleHash) {
    simpleHash.fallbackSimpleHash = function (arrayBuffer) {
      const bytes = new Uint8Array(arrayBuffer);
      let finalHexString = "";
      const sz = [0x1234, 0x5678, 0x9abc, 0xdef0, 0x2468];
      for (let i = 0; i < bytes.length; i++) {
        const index = i % 5;
        sz[index] = sz[index] * 131 + bytes[i];
        sz[index] |= 0;
      }
      for (const hash of sz) {
        finalHexString += (hash >>> 0).toString(16).padStart(8, "0");
      }
      return finalHexString;
    };
  }
  const { numberOfSamples = 200, sampleSize = 1024 } = options;
  const samples = await simpleHash.smartFileSampler(
    file,
    numberOfSamples,
    sampleSize,
  );
  const totalLength = samples.reduce((sum, buf) => sum + buf.byteLength, 0);
  const combinedBuffer = new Uint8Array(totalLength);
  let offset = 0;
  for (const sample of samples) {
    combinedBuffer.set(new Uint8Array(sample), offset);
    offset += sample.byteLength;
  }
  try {
    const hashBuffer = await crypto.subtle.digest("SHA-1", combinedBuffer);
    const hashArray = Array.from(new Uint8Array(hashBuffer));
    const hashHex = hashArray
      .map((b) => b.toString(16).padStart(2, "0"))
      .join("");
    return hashHex;
  } catch (error) {
    return simpleHash.fallbackSimpleHash(combinedBuffer);
  }
};
