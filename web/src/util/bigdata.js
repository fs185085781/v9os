const allData = {}
function initBigDb(fn) {
    if (allData.indexDb) {
        fn(true);
        return;
    }
    const indexedDB = window.indexedDB || window.mozIndexedDB || window.webkitIndexedDB || window.msIndexedDB;
    try {
        let req = indexedDB.open("v9os");
        req.onerror = function (event) {
            if (fn) {
                fn(false);
            }
        };
        req.onsuccess = function (e) {
            allData.indexDb = e.target.result;
            if (fn) {
                fn(true);
            }
        };
        req.onupgradeneeded = function (e) {
            e.target.result.createObjectStore('cacheData', {
                keyPath: 'key'
            });
        }
    } catch (e) {
        if (fn) {
            fn(false);
        }
    }
}
export function getBigData(key) {
    return new Promise(function (success) {
        initBigDb(function (flag) {
            if (!flag) {
                success(false);
                return;
            }
            let req = allData.indexDb.transaction(["cacheData"]).objectStore("cacheData").get(key);
            req.onerror = function (event) {
                success(false);
            };
            req.onsuccess = function (event) {
                let data = undefined;
                if (req.result !== undefined) {
                    data = req.result.val;
                }
                success(data)
            };
        });
    });

}
export async function setBigData(key, val) {
    val = JSON.parse(JSON.stringify(val))
    let data = await getBigData(key);
    if (data === false) {
        return false;
    }
    return new Promise(function (success) {
        let req;
        if (data === undefined) {
            //新增
            req = allData.indexDb.transaction(['cacheData'], 'readwrite')
                .objectStore('cacheData')
                .add({ key: key, val: val });
        } else {
            //修改
            req = allData.indexDb.transaction(['cacheData'], 'readwrite')
                .objectStore('cacheData')
                .put({ key: key, val: val });
        }
        req.onsuccess = function (event) {
            success(true)
        };
        req.onerror = function (event) {
            success(false)
        }
    });
}
export function delBigData(key) {
    return new Promise(function (success) {
        initBigDb(function (flag) {
            if (!flag) {
                success(false);
                return;
            }
            let req = that.indexDb.transaction(["cacheData"], "readwrite").objectStore("cacheData").delete(key);
            req.onerror = function (event) {
                success(false);
            };
            req.onsuccess = function (event) {
                success(true);
            };
        });
    });

}
