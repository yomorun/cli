
function handler(data) {
    // log.Println(">> JS Handler Begin");
    // to string
    decodedString = arrayBufferToString(data);
    // parse JSON
    var value = JSON.parse(decodedString);
    value.from = "JS SFN处理>" + value.from
    value.noise = value.noise / 10
    // log.Printf(">> data.JSON: %v, type:%v", value, typeof value);

    payload = JSON.stringify(value)
    buf = stringToArrayBuffer(payload)

    rightNow = Date.now()
    log.Printf("[stream-js-fn] from=%s, Timestamp=%d, value=%f (⚡️=%dms)", value.from, value.time, value.noise, rightNow - value.time)
    return {"id": 0x34, "data": buf}
}

function dataID() {
    return [0x33]
}

