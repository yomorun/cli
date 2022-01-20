
function handler(data) {
    // log.Println(">> JS Handler Begin");
    var uint8buf = new Uint8Array(data);
    // to string
    // var decodedString = decode(uint8buf);
    decodedString = arrayBufferToString(data);
    // log.Printf(">> data.string: %s", decodedString);
    // var noise float32
    // err = json.Unmarshal(payload, &noise)
    // if err != nil {
    // 	log.Printf(">> [sink] unmarshal data failed, err=%v", err)
    // } else {
    // 	log.Printf(">> [sink](Compile) save `%v` to FaunaDB\n", noise)
    // }

    // result
    // log.Printf(">> [sink](Compile) return id `%v`\n", id)
    // parse JSON
    var jsonData = JSON.parse(decodedString);
    jsonData.from = "sink处理>" + jsonData.from
    log.Printf(">> data.JSON: %v, type:%v", jsonData, typeof jsonData);
    log.Printf(">> data.Hex: %v", ab2hex(data));
    // log.Println(">> JS Handler End");
    return {"id": 0x35, "data": null}
}

function dataID() {
    return [0x34]
}

function decode(uint8buf) {
    var encodedString = String.fromCharCode.apply(null, uint8buf);
    var decodedString = decodeURIComponent(escape(encodedString));
    return decodedString;
}

// Uint8Array to Hex
function ab2hex(buffer) {
    var hexArr = Array.prototype.map.call(
        new Uint8Array(buffer),
        function (bit) {return ('00' + bit.toString(16)).slice(-2)}
    )
    return hexArr.join('')
}
