
function handler(data) {
    log.Println(">> JS Handler Begin");
    var uint8buf = new Uint8Array(data);
    // to string
    decodedString = decode(uint8buf);
    log.Printf(">> data.string: %s", decodedString);
    // parse JSON
    var jsonData = JSON.parse(decodedString);
    log.Printf(">> data.JSON: %v", jsonData);
    log.Printf(">> data.Hex: %v", ab2hex(data));
    log.Println(">> JS Handler End");
    return {id: 0x35, data: data}
}

function dataID() {
    return [52, 53]
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
