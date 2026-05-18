(function () {
  let src = document.currentScript.src;
  let sz = src.split("/");
  sz.length -= 1;
  let parent = sz.join("/");
  document.writeln(`<script src="${parent}/manifest.js"></script>`);
  document.writeln(`<script src="${parent}/core.js"></script>`);
  sz.length -= 1;
  window.__APP_WEB_BASE_PATH = sz.join("/");
})();
