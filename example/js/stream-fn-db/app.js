
function handler(data) {
    var uint8buf = new Uint8Array(data);
    // to string
    var decodedString = decode(uint8buf);
    // parse JSON
    var value = JSON.parse(decodedString);
    value.from = value.from + ">JS SINK";
    log.Printf(">> [sink] save %v to FaunaDB", value);
}

function dataTags() {
    return [0x34]
}

function decode(uint8buf) {
    var encodedString = String.fromCharCode.apply(null, uint8buf);
    var decodedString = decodeURIComponent(escape(encodedString));
    return decodedString;
}
