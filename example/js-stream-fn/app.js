// function Handler(data []byte) (byte, []byte) {
function Handler(data) {
    // var noise = 0f
    // err:= json.Unmarshal(data, & noise)
    // if err != nil {
    //     log.Printf(">> [sink] unmarshal data failed, err=%v", err)
    // } else {
    //     log.Printf(">> [sink] save `%v` to FaunaDB\n", noise)
    // }

    // return 0x0, nil
    return 0x0, data
}

// function DataID() []byte {
function DataID() {
    // return []byte {0x34}
    // return Uint8Array.of(0x34)
    return [0x34]
}
