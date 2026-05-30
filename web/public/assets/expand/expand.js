(function (global) {
  var $ = global.jQuery || global.$;
  if (!$) return;

  if (!$.fn.die) {
    $.fn.die = function (types, selector) {
      return this.off(types, selector);
    };
  }

  if (!$.fn.inScreen) {
    $.fn.inScreen = function () {
      var elem = this[0];
      if (!elem) return false;
      var rect = elem.getBoundingClientRect();
      var width = window.innerWidth || document.documentElement.clientWidth;
      var height = window.innerHeight || document.documentElement.clientHeight;
      return rect.bottom >= 0 && rect.right >= 0 && rect.top <= height && rect.left <= width;
    };
  }

  if (!$.setStyle) {
    $.setStyle = function (cssText, id) {
      var head = document.getElementsByTagName("head")[0] || document.documentElement;
      var element = document.getElementById(id);
      $(element).remove();
      if (!cssText) return;
      element = document.createElement("style");
      if (id) element.id = id;
      element.type = "text/css";
      head.appendChild(element);
      if (element.styleSheet !== undefined) {
        if (31 < document.getElementsByTagName("style").length) {
          throw new Error("Exceed the maximal count of style tags in IE");
        }
        element.styleSheet.cssText = cssText;
      } else {
        element.appendChild(document.createTextNode(cssText));
      }
    };
  }

  if ($.ajaxTransport) {
    $.ajaxTransport("+binary", function (options, originalOptions, jqXHR) {
      if (
        window.FormData &&
        window.ArrayBuffer &&
        window.Blob &&
        "binary" == options.dataType &&
        (!options.data || options.data instanceof ArrayBuffer || options.data instanceof Blob)
      ) {
        return {
          send: function (headers, callback) {
            var i,
              xhr = options.xhr ? options.xhr() : new XMLHttpRequest(),
              url = options.url,
              type = options.type,
              async = options["async"] || true,
              dataType = options.responseType || "blob",
              data = options.data || null,
              username = options.username || null,
              password = options.password || null;
            for (
              i in xhr.addEventListener("load", function () {
                var data = {};
                data[options.dataType] = xhr.response;
                callback(xhr.status, xhr.statusText, data, xhr.getAllResponseHeaders());
              }),
              xhr.addEventListener("error", function () {
                var data = {};
                data[options.dataType] = xhr.response;
                callback(xhr.status, xhr.statusText, data, xhr.getAllResponseHeaders());
              }),
              xhr.open(type, url, async, username, password),
              headers
            ) {
              xhr.setRequestHeader(i, headers[i]);
            }
            xhr.responseType = dataType;
            xhr.send(data);
          },
          abort: function () {
            jqXHR.abort();
          }
        };
      }
    });
  }
  if (!global.htmlEncode) {
    global.htmlEncode = function (str, notSpace) {
      var s = (str || "") + "";
      return 0 === str ? "0" : str && s && 0 != s.length ? (s = (s = (s = (s = s.replace(/&/g, "&amp;")).replace(/</g, "&lt;")).replace(/>/g, "&gt;")).replace(/ /g, " "),
        (s = (s = notSpace ? s : s.replace(/ /g, "&nbsp;")).replace(/\'/g, "&#39;")).replace(/\"/g, "&quot;")) : ""
    }
  }
})(window);
