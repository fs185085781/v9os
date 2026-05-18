(function () {
  //注入manifest.json
  let str = localStorage.getItem("mainfestData");
  if (str) {
    let mainfest = JSON.parse(str);
    if (mainfest) {
      let data = {
        short_name: mainfest.name,
        name: mainfest.name,
        icons: [
          { src: mainfest.logo, type: "image/png", sizes: mainfest.size },
        ],
        start_url: window.location.origin,
        display: "standalone",
      };
      let mainfestUrl = URL.createObjectURL(
        new Blob([JSON.stringify(data)], { type: "application/json" }),
      );
      document.writeln('<link rel="manifest" href="' + mainfestUrl + '"/>');
      document.writeln(
        '<link rel="shortcut icon" href="' +
          mainfest.logo +
          '" type="image/x-icon">',
      );
    }
  }
  const getImageSize = async (url) => {
    return new Promise((resolve, reject) => {
      const img = new Image();
      img.onload = () => {
        resolve({
          width: img.naturalWidth,
          height: img.naturalHeight
        });
      };
      img.onerror = reject;
      img.src = url;
    });
  }
  window.__MAIN_FEST_SAVE = async function (data) {
    if (!data) {
      return;
    }
    let size = await getImageSize(data.logo);
    if(size.width > 0 && size.height > 0){
      size = size.width + "x" + size.height;
    }else{
      size = "192x192";
    }
    data.size = size;
    localStorage.setItem("mainfestData", JSON.stringify(data));
  };
})();
