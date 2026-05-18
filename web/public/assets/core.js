(function () {
  /**
   * 1.使用serviceWorker支持离线模式
   * 2.高性能hash计算
   * 3.封装时间处理,包含时间转化,时间格式化,时间加减,时间求差
   * 4.封装数值处理,包含数值格式化,数值精准计算
   */
  /**serviceWorker拓展----开始*/
  if ("serviceWorker" in navigator) {
    const sw = {
      commonMessage: function (qid, type, data) {
        return new Promise((resolve) => {
          const channel = new MessageChannel();
          channel.port1.onmessage = (event) => {
            if (event.data.qid && event.data.qid === qid) {
              resolve(event.data.data);
            } else {
              resolve(false);
            }
          };
          navigator.serviceWorker.controller.postMessage(
            {
              type,
              data,
              qid,
            },
            [channel.port2],
          );
        });
      },
      getMode: function () {
        return this.commonMessage("GET_MODE", "GET_MODE");
      },
    };
    navigator.serviceWorker.addEventListener("controllerchange", () => {
      if (navigator.serviceWorker.controller) {
        window.sw = sw;
      }
    });
    navigator.serviceWorker.addEventListener("message", function (event) {
      window.postMessage(event.data, "*");
    });
    //必须使用/sw.js才能拦截全部请求
    navigator.serviceWorker
      .register("/sw.js")
      .then(function () {
        return navigator.serviceWorker.ready;
      })
      .then(function () {
        if (navigator.serviceWorker.controller) {
          window.sw = sw;
        }
      })
      .catch(function () {});
  }
  /**hash计算拓展----开始*/
  let src = document.currentScript.src;
  let sz = src.split("/");
  sz.length -= 1;
  document.writeln(`<script src="${sz.join("/")}/simplehash.js"></script>`);
  /**时间拓展----开始*/
  function numToMonth(m, flag) {
    function numToStr(i) {
      switch (i) {
        case 1:
          return "一";
        case 2:
          return "二";
        case 3:
          return "三";
        case 4:
          return "四";
        case 5:
          return "五";
        case 6:
          return "六";
        case 7:
          return "七";
        case 8:
          return "八";
        case 9:
          return "九";
        case 10:
          return "十";
        case 11:
          return "十一";
        case 12:
          return "十二";
      }
      return "NaN";
    }
    if (flag) {
      return numToStr(m) + "月";
    } else {
      return m + "月";
    }
  }
  function numToWeek(w, flag) {
    if (flag == 3) {
      function numToStr(i) {
        switch (i) {
          case 0:
            return "Sun";
          case 1:
            return "Mon";
          case 2:
            return "Tue";
          case 3:
            return "Wed";
          case 4:
            return "Thu";
          case 5:
            return "Fri";
          case 6:
            return "Sat";
        }
        return "NaN";
      }
      return numToStr(w);
    }
    function numToStr(i) {
      switch (i) {
        case 0:
          return "日";
        case 1:
          return "一";
        case 2:
          return "二";
        case 3:
          return "三";
        case 4:
          return "四";
        case 5:
          return "五";
        case 6:
          return "六";
      }
      return "NaN";
    }
    let r = numToStr(w);
    if (flag == 1) {
      r = "星期" + r;
    } else if (flag == 2) {
      r = "周" + r;
    }
    return r;
  }
  function numToQuarter(m, flag) {
    if (m > 9) {
      return flag ? "四季度" : "Q4";
    } else if (m > 6) {
      return flag ? "三季度" : "Q3";
    } else if (m > 3) {
      return flag ? "二季度" : "Q2";
    } else {
      return flag ? "一季度" : "Q1";
    }
  }
  function numToHealthyYear(m) {
    return m > 6 ? "下半年" : "上半年";
  }
  let dateUtils = {
    formatDate: function (date, format) {
      if (!date || !(date instanceof Date)) {
        return "";
      }
      if (!format) {
        format = "yyyy-MM-dd HH:mm:ss";
      }
      let yyyy = date.getFullYear();
      let data = {
        y: yyyy - parseInt(yyyy / 100) * 100,
        M: date.getMonth() + 1,
        d: date.getDate(),
        H: date.getHours(),
        h: date.getHours() > 12 ? date.getHours() - 12 : date.getHours(),
        m: date.getMinutes(),
        s: date.getSeconds(),
        f: date.getMilliseconds(),
      };
      let sfs = ["y", "M", "d", "H", "h", "m", "s", "f"];
      for (let i = 0; i < sfs.length; i++) {
        let key = sfs[i];
        data[key + key] = data[key] < 10 ? "0" + data[key] : "" + data[key];
      }
      data.MMM = numToMonth(data.M);
      data.MMMM = numToMonth(data.M, true);
      data.ddd = numToWeek(date.getDay());
      data.dddd = numToWeek(date.getDay(), 1);
      data.ddddd = numToWeek(date.getDay(), 2);
      data.dddddd = numToWeek(date.getDay(), 3);
      data.yyyy = yyyy;
      data.fff =
        data.f < 100 ? (data.f < 10 ? "00" + data.f : "0" + data.f) : data.f;
      data.tt = data.H >= 12 ? "下午" : "上午";
      data.t = data.H >= 12 ? "PM" : "AM";
      data.Q = numToQuarter(data.M, true);
      data.q = numToQuarter(data.M);
      data.hy = numToHealthyYear(data.M);
      let fields = [
        "yyyy",
        "yy",
        "hy",
        "y",
        "Q",
        "q",
        "MMMM",
        "MMM",
        "MM",
        "M",
        "dddddd",
        "ddddd",
        "dddd",
        "ddd",
        "dd",
        "d",
        "HH",
        "H",
        "hh",
        "h",
        "mm",
        "m",
        "ss",
        "s",
        "fff",
        "ff",
        "f",
        "tt",
        "t",
      ];
      for (let i = 0; i < fields.length; i++) {
        let field = fields[i];
        format = format.replace(new RegExp(field, "g"), data[field]);
      }
      return format;
    },
    parseDate: function (str, ignoreTimeZone) {
      try {
        let date = eval(str);
        if (date && date.getFullYear) {
          return date;
        }
      } catch (e) {}
      if (typeof str == "object") {
        return isNaN(str) ? null : str;
      }
      if (typeof str == "number") {
        let date = new Date(str);
        if (date.getTime() != str) {
          return null;
        }
        return isNaN(date) ? null : date;
      }
      if (typeof str == "string") {
        let m = str.match(/^([0-9]{4})([0-9]{2})([0-9]{0,2})$/);
        if (m) {
          let date = new Date(parseInt(m[1], 10), parseInt(m[2], 10) - 1);
          if (m[3]) {
            date.setDate(m[3]);
          }
          return date;
        }
        m = str.match(/^([0-9]{4}).([0-9]*)$/);
        if (m) {
          let date = new Date(m[1], m[2] - 1);
          return date;
        }
        if (str.match(/^\d+(\.\d+)?$/)) {
          let ms = parseInt(str);
          let date = new Date(ms);
          if (date.getTime() != ms) {
            return null;
          }
          return isNaN(date) ? null : date;
        }
        if (ignoreTimeZone === undefined) {
          ignoreTimeZone = true;
        }
        let date =
          this.parseISO8601(str, ignoreTimeZone) ||
          (str ? new Date(str) : null);
        return isNaN(date) ? null : date;
      }
      return null;
    },
    parseISO8601: function (str, ignoreTimeZone) {
      let m = str.match(
        /^([0-9]{4})([-\/]([0-9]{1,2})([-\/]([0-9]{1,2})([T ]([0-9]{1,2}):([0-9]{1,2})(:([0-9]{1,2})(\.([0-9]+))?)?(Z|(([-+])([0-9]{2})(:?([0-9]{2}))?))?)?)?)?$/,
      );
      if (!m) {
        m = str.match(
          /^([0-9]{4})[-\/]([0-9]{2})[-\/]([0-9]{2})[T ]([0-9]{1,2})/,
        );
        if (m) {
          let date = new Date(m[1], m[2] - 1, m[3], m[4]);
          return date;
        }
        m = str.match(/^([0-9]{4}).([0-9]*)$/);
        if (m) {
          let date = new Date(m[1], m[2] - 1);
          return date;
        }
        m = str.match(/^([0-9]{4}).([0-9]*).([0-9]*)/);
        if (m) {
          let date = new Date(m[1], m[2] - 1, m[3]);
          return date;
        }
        m = str.match(/^([0-9]{2})-([0-9]{2})-([0-9]{4})$/);
        if (!m) {
          return null;
        } else {
          let date = new Date(m[3], m[1] - 1, m[2]);
          return date;
        }
      }
      let date = new Date(m[1], 0, 1);
      if (ignoreTimeZone || !m[14]) {
        let tempDate = new Date(m[1], 0, 1, 9, 0);
        if (m[3]) {
          date.setMonth(m[3] - 1);
          tempDate.setMonth(m[3] - 1);
        }
        if (m[5]) {
          date.setDate(m[5]);
          tempDate.setDate(m[5]);
        }
        this.fixDate(date, tempDate);
        if (m[7]) {
          date.setHours(m[7]);
        }
        if (m[8]) {
          date.setMinutes(m[8]);
        }
        if (m[10]) {
          date.setSeconds(m[10]);
        }
        if (m[12]) {
          date.setMilliseconds(Number("0." + m[12]) * 1000);
        }
        this.fixDate(date, tempDate);
      } else {
        date.setUTCFullYear(m[1], m[3] ? m[3] - 1 : 0, m[5] || 1);
        date.setUTCHours(
          m[7] || 0,
          m[8] || 0,
          m[10] || 0,
          m[12] ? Number("0." + m[12]) * 1000 : 0,
        );
        let num = Number(m[16]) * 60 + (m[18] ? Number(m[18]) : 0);
        num *= m[15] == "-" ? 1 : -1;
        date = new Date(+date + num * 60 * 1000);
      }
      return date;
    },
    fixDate: function (date, tempDate) {
      if (!+date) {
        return;
      }
      while (date.getDate() != tempDate.getDate()) {
        date.setTime(+date + (date < tempDate ? 1 : -1) * 3600000);
      }
    },
  };
  Date.prototype.addMilliseconds = function (x) {
    let v = this.getTime() + x;
    return new Date(v);
  };
  Date.prototype.addMonths = function (x) {
    let months = this.getFullYear() * 12 + this.getMonth() + x + 1;
    let date = this.getDate();
    let year = parseInt(months / 12);
    let temp = new Date(
      year,
      months - year * 12,
      1,
      this.getHours(),
      this.getMinutes(),
      this.getSeconds(),
      this.getMilliseconds(),
    );
    let data = temp.addDays(-1);
    if (date < data.getDate()) {
      data.setDate(date);
    }
    return data;
  };
  let dateMethods = [
    { prop: "addSeconds", parentProp: "addMilliseconds", pow: 1000 },
    { prop: "addMinutes", parentProp: "addSeconds", pow: 60 },
    { prop: "addHours", parentProp: "addMinutes", pow: 60 },
    { prop: "addDays", parentProp: "addHours", pow: 24 },
    { prop: "addYears", parentProp: "addMonths", pow: 12 },
  ];
  for (let i = 0; i < dateMethods.length; i++) {
    let temp = dateMethods[i];
    (function (map) {
      Date.prototype[map.prop] = function (x) {
        return this[map.parentProp](x * map.pow);
      };
    })(temp);
  }
  let betweenMethods = [
    { prop: "betweenMilliseconds", pow: 1 },
    { prop: "betweenSeconds", pow: 1000 },
    { prop: "betweenMinutes", pow: 1000 * 60 },
    { prop: "betweenHours", pow: 1000 * 60 * 60 },
    { prop: "betweenDays", pow: 1000 * 60 * 60 * 24 },
  ];
  /**根据回调获取差值*/
  function getCha(d1, d2, callback) {
    let max, min;
    if (d1.getTime() >= d2.getTime()) {
      max = d1;
      min = d2;
    } else {
      max = d2;
      min = d1;
    }
    let cha = callback(max, min);
    return { max: max, min: min, cha: cha };
  }
  /**比较日期+小时+分钟+秒+毫秒的大小*/
  function isMax(d1, d2) {
    let max, min;
    if (d1.getTime() >= d2.getTime()) {
      max = d1;
      min = d2;
    } else {
      max = d2;
      min = d1;
    }
    let xms =
      max.getDate() * 24 * 60 * 60 * 1000 +
      max.getHours() * 60 * 60 * 1000 +
      max.getMinutes() * 60 * 1000 +
      max.getSeconds() * 1000 +
      max.getMilliseconds();
    let tms =
      min.getDate() * 24 * 60 * 60 * 1000 +
      min.getHours() * 60 * 60 * 1000 +
      min.getMinutes() * 60 * 1000 +
      min.getSeconds() * 1000 +
      min.getMilliseconds();
    return xms >= tms;
  }
  for (let i = 0; i < betweenMethods.length; i++) {
    let temp = betweenMethods[i];
    (function (map) {
      Date.prototype[map.prop] = function (x) {
        let options = getCha(this, x, function (max, min) {
          return max.getTime() - min.getTime();
        });
        return parseInt(options.cha / map.pow);
      };
    })(temp);
  }
  Date.prototype.betweenMonths = function (x) {
    let options = getCha(this, x, function (max, min) {
      return (
        max.getFullYear() * 12 +
        max.getMonth() -
        min.getFullYear() * 12 -
        min.getMonth()
      );
    });
    if (isMax(options.min, options.max)) {
      return options.cha;
    }
    return options.cha - 1;
  };
  Date.prototype.betweenYears = function (x) {
    let options = getCha(this, x, function (max, min) {
      return max.getFullYear() - min.getFullYear();
    });
    if (options.max.getMonth() > options.min.getMonth()) {
      return options.cha;
    }
    if (options.max.getMonth() < options.min.getMonth()) {
      return options.cha - 1;
    }
    if (isMax(options.min, options.max)) {
      return options.cha;
    }
    return options.cha - 1;
  };
  Date.prototype.format = function (format) {
    return dateUtils.formatDate(this, format);
  };
  Date.parseDate = function (str, ignoreTimeZone) {
    if (Date.parse) {
      let number = Date.parse(str);
      if (!isNaN(number)) {
        return new Date(number);
      }
    }
    return dateUtils.parseDate(str, ignoreTimeZone);
  };
  /**时间拓展----结束*/
  /**数值拓展----开始*/
  let numberUtils = {
    formatNumber: function (number, format) {
      function joinZeroStr(nums) {
        let str = "";
        while (nums > 0) {
          nums -= 1;
          str += "0";
        }
        return str;
      }
      function fillZeroPosition(value, fixedNums, isRightFill) {
        value = value + "";
        fixedNums = typeof fixedNums == "number" ? fixedNums : 2;
        let fillNums = fixedNums - value.length;
        if (fillNums > 0) {
          let str = joinZeroStr(fillNums);
          return isRightFill ? value + str : str + value;
        }
        return value;
      }
      function getFormatValue(value, format, formatMap) {
        let regExp = /[^0#]/g;
        value = Math.abs(value);
        let hasDh = format.indexOf(",") != -1;
        let formatSplit = format.split(".");
        let integralFormat = (formatSplit[0] || "").replace(regExp, "");
        let fractionalFormat = (formatSplit[1] || "").replace(regExp, "");
        let formatValue = "";
        let decimalsSeparator = formatMap.decimalsSeparator;
        let groupSeparator = formatMap.groupSeparator;
        let hasZero = integralFormat.indexOf("0");
        integralFormat =
          hasZero == -1 ? "0" : integralFormat.substr(hasZero) || "0";
        let fractionalLength = fractionalFormat.length;
        let fillZreoLenth = fractionalFormat.substr(
          0,
          fractionalFormat.lastIndexOf("0") + 1,
        ).length;
        function numRound(number, fractionDigits) {
          return (
            Math.round(number * Math.pow(10, fractionDigits)) /
            Math.pow(10, fractionDigits)
          );
        }
        value = numRound(value, fractionalLength);
        let valueSz = String(value).split(".");
        let integralValue = valueSz[0];
        let fractionalValue = valueSz[1] || "";
        if (integralValue) {
          integralValue = fillZeroPosition(
            integralValue,
            integralFormat.length,
          );
          if (hasDh) {
            for (
              let i = 0;
              i < Math.floor((integralValue.length - (1 + i)) / 3);
              i++
            ) {
              integralValue =
                integralValue.substring(0, integralValue.length - (4 * i + 3)) +
                groupSeparator +
                integralValue.substring(integralValue.length - (4 * i + 3));
            }
          }
          formatValue += integralValue;
        }
        if (fractionalLength > 0) {
          fractionalValue = fillZeroPosition(
            fractionalValue.substr(0, fractionalLength),
            fillZreoLenth,
            true,
          );
          if (fractionalValue !== "") {
            formatValue += decimalsSeparator;
            formatValue += fractionalValue;
          }
        }
        return formatValue;
      }
      function getFormatIndex(format) {
        if (!format) {
          return null;
        }
        function getZeroIndexs(format) {
          let zeroIndex = format.indexOf("0");
          let hashIndex = format.indexOf("#");
          if (zeroIndex == -1 || (hashIndex != -1 && hashIndex < zeroIndex)) {
            zeroIndex = hashIndex;
          }
          let lastZeroIndex = format.lastIndexOf("0");
          let lastHashIndex = format.lastIndexOf("#");
          if (
            lastZeroIndex == -1 ||
            (lastHashIndex != -1 && lastHashIndex > lastZeroIndex)
          ) {
            lastZeroIndex = lastHashIndex;
          }
          return [zeroIndex, lastZeroIndex];
        }
        let zeroIndexs = getZeroIndexs(format);
        let zeroIndex = zeroIndexs[0];
        let lastZeroIndex = zeroIndexs[1];
        return zeroIndex > -1
          ? {
              begin: zeroIndex,
              end: lastZeroIndex,
              format: format.substring(zeroIndex, lastZeroIndex + 1),
            }
          : null;
      }
      function hasZeroOrHash(value) {
        return value.indexOf("0") != -1 || value.indexOf("#") != -1;
      }
      function getCommonFormat(format, number, formatMap, formatSz) {
        let ncpReg = /^(n|c|p)(\d*)$/i;
        let formatAttrs = formatMap.number;
        let ncpRes = ncpReg.exec(format);
        if (ncpRes != null) {
          let ncpValue = ncpRes[1];
          let ncpCount = ncpRes[2];
          if (ncpValue == "p") {
            formatAttrs = formatMap.percent;
          } else if (ncpValue == "c") {
            formatAttrs = formatMap.currency;
          }
          let pointCount = ncpCount ? parseInt(ncpCount) : formatAttrs.decimals;
          let plusMinusFormat = formatAttrs.pattern[number < 0 ? 1 : 0];
          plusMinusFormat = plusMinusFormat.replace(
            "n",
            "#,#" + (pointCount > 0 ? "." + joinZeroStr(pointCount) : ""),
          );
          format = format
            .replace(ncpValue + ncpCount, plusMinusFormat)
            .replace("$", formatMap.currency.symbol)
            .replace("%", formatMap.percent.symbol);
        } else if (hasZeroOrHash(format)) {
          if (number < 0 && !formatSz[1]) {
            format = "-" + format;
          }
        }
        return format;
      }
      let indexNumberReg = /^(e)(\d*)$/i;
      if (typeof number != "number") {
        return "";
      }
      if (!format) {
        return String(number);
      }
      let formatSz = format.split(";");
      format = formatSz[0];
      if (number < 0 && formatSz.length >= 2) {
        format = formatSz[1];
      }
      if (number == 0 && formatSz.length >= 3) {
        format = formatSz[2];
      }
      let formatMap = {
        number: {
          pattern: ["n", "-n"],
          decimals: 2,
          decimalsSeparator: ".",
          groupSeparator: ",",
          groupSize: [3],
        },
        percent: {
          pattern: ["n%", "-n%"],
          decimals: 2,
          decimalsSeparator: ".",
          groupSeparator: ",",
          groupSize: [3],
          symbol: "%",
        },
        currency: {
          pattern: ["$n", "$-n"],
          decimals: 2,
          decimalsSeparator: ".",
          groupSeparator: ",",
          groupSize: [3],
          symbol: "¥",
        },
      };
      let numberFormat = formatMap.number;
      let percentFormat = formatMap.percent;
      let currencyFormat = formatMap.currency;
      let commonFormat = getCommonFormat(format, number, formatMap, formatSz);
      let hasCurrencySymbol = commonFormat.indexOf(currencyFormat.symbol) != -1;
      let hasPercentSymbol = commonFormat.indexOf(percentFormat.symbol) != -1;
      let hasZeroHash = hasZeroOrHash(commonFormat);
      let realFormat = hasCurrencySymbol
        ? currencyFormat
        : hasPercentSymbol
          ? currencyFormat
          : numberFormat;
      let formatNumber = hasPercentSymbol ? number * 100 : number;
      let indexNumberRes = indexNumberReg.exec(commonFormat);
      if (indexNumberRes) {
        let indexCount = parseInt(indexNumberRes[2]);
        return isNaN(indexCount)
          ? formatNumber.toExponential()
          : formatNumber.toExponential(indexCount);
      }
      if (!hasZeroHash) {
        return commonFormat;
      }
      let formatValue = "";
      let formatIndex = getFormatIndex(commonFormat);
      if (formatIndex != null) {
        formatValue = getFormatValue(
          formatNumber,
          formatIndex.format,
          realFormat,
        );
        formatValue =
          commonFormat.substr(0, formatIndex.begin) +
          formatValue +
          commonFormat.substr(formatIndex.end + 1);
      } else {
        formatValue = commonFormat;
      }
      return formatValue;
    },
  };
  Number.prototype.format = function (format) {
    return numberUtils.formatNumber(this.valueOf(), format);
  };
  /**数值拓展----结束*/
  /**精准计算拓展---开始*/
  let numberMethods = ["floor", "round", "ceil"];
  for (let i = 0; i < numberMethods.length; i++) {
    let field = numberMethods[i];
    (function (f) {
      Number.prototype[f] = function (x) {
        if (x < 0) {
          x = 0;
        }
        let pow = Math.pow(10, x);
        let temp = Math[f](this * pow);
        return temp / pow;
      };
    })(field);
  }
  let decimalMethods = ["add", "subtract", "multiply", "divide"];
  for (let i = 0; i < decimalMethods.length; i++) {
    let field = decimalMethods[i];
    (function (f) {
      Number.prototype[f] = function (x, maxJd, type) {
        if (isNaN(x)) {
          throw x + " is not a number";
        }
        if (typeof x != "number") {
          x = x * 1;
        }
        if (isNaN(maxJd)) {
          console.warn(maxJd + " is not a number,will be reset to 16");
          maxJd = 16;
        }
        if ((!maxJd && maxJd != 0) || maxJd > 16) {
          maxJd = 16;
        }
        type = type || "round";
        if (type != "floor" && type != "round" && type != "ceil") {
          console.warn(type + " is not a number,will be reset to 'round'");
          type = "round";
        }
        function getAssist(v) {
          let str = String(v);
          let eindex = str.indexOf("e");
          let addres = 0;
          if (eindex != -1) {
            addres = str.substring(eindex + 1) * -1;
            str = str.substring(0, eindex);
          }
          let index = str.indexOf(".");
          let res = index != -1 ? str.length - index - 1 : 0;
          res = res + addres;
          if (res > maxJd) {
            res = maxJd;
          }
          if (res < 0) {
            res = 0;
          }
          let pow = Math.pow(10, res);
          let num = v[type](res);
          return { pow: pow, num: num };
        }
        let a = getAssist(this);
        let b = getAssist(x);
        let max = a.pow > b.pow ? a.pow : b.pow;
        let jg = NaN;
        if (f == "add") {
          jg = (a.num * max + b.num * max) / max;
        } else if (f == "subtract") {
          jg = (a.num * max - b.num * max) / max;
        } else if (f == "multiply") {
          jg = (a.num * a.pow * (b.num * b.pow)) / (a.pow * b.pow);
        } else if (f == "divide") {
          jg = (a.num * max) / (b.num * max);
        }
        return jg;
      };
    })(field);
  }
  /**精准计算拓展---结束*/
})();
