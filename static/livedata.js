const EndDeviceLiveTable = ({ type, liveDataInfo }) => {
    const [data, setData] = React.useState([])
    const [webSocket, setWebSocket] = React.useState(null)
    React.useEffect(() => {
        if (webSocket)
        {
            webSocket.onmessage = (e) => {
                data.unshift(JSON.parse(e.data))
                setData(data.slice(0, 10))
            }
        }
    }, [data])
    
    React.useEffect(() => {
        if (type == DataType.EndDevices)
        {
            let path = liveDataInfo.path.replace("http", "ws") + "/live"
            let ws = new WebSocket(path)
            setWebSocket(ws)

            path = liveDataInfo.path + "/activity"
            axios.get(path).then(res => {
                setData(res.data)
            })
            return () => {
                if (ws)
                {
                    ws.close()
                    ws = null;
                }
            }
        }
    }, [])

    const formatBusId = (devEui) => {
        if (typeof devEui !== "bigint") return
        return (devEui & BigInt(0xFFFFFF)).toString()
    }

    const liveData = data.map((res, index) => { 
        let date = new Date(res.Time)
        return (
        <tr>
            <th scope="row">{index + 1}</th>
            <td>{date.toLocaleDateString()} {date.toTimeString().split(' ')[0]}</td>
            <td>{GetFrameType(res.FType)}</td>
            <td>{res.Payload}</td>
            <td>{atob(res.Payload)}</td>
            <td>{base64ToHexArray(res.Payload)}</td>
        </tr>
        )
    })

    return (
        type == DataType.EndDevices
        ?
        <>
        <h1>Bus {formatBusId(liveDataInfo.uniqueId)}</h1>
        <table class="table">
            <thead>
                <tr>
                <th scope="col">#</th>
                <th scope="col">Time</th>
                <th scope="col">Type</th>
                <th scope="col">Payload (bytes)</th>
                <th scope="col">Payload (string)</th>
                <th scope="col">Payload (hex)</th>
                </tr>
            </thead>
            <tbody>{ liveData }</tbody>
        </table>
        </>
        :
        <></>
    )
}

const GatewayLiveTable = ({ type, liveDataInfo }) => {
    const [data, setData] = React.useState([])
    const [webSocket, setWebSocket] = React.useState(null)
    React.useEffect(() => {
        if (webSocket)
        {
            webSocket.onmessage = (e) => {
                data.unshift(JSON.parse(e.data))
                setData(data.slice(0, 10))
            }
        }
    }, [data])

    React.useEffect(() => {
        if (type == DataType.Gateway)
        {
            let path = liveDataInfo.path.replace("http", "ws") + "/live"
            let ws = new WebSocket(path)
            setWebSocket(ws)

            path = liveDataInfo.path + "/activity"
            axios.get(path).then(res => {
                setData([...res.data])
            })
            return () => {
                if (ws)
                {
                    ws.close()
                    ws = null;
                }
            }
        }
    }, [])

    const liveData = data.map((res, index) => { 
        let date = new Date(res.Time)
        return (
            <tr>
                <th scope="row">{index + 1}</th>
                <td>{date.toLocaleDateString()} {date.toTimeString().split(' ')[0]}</td>
                <td>{GetFrameType(res.FType)}</td>
                <td>{res.Rssi}</td>
                <td>{res.Snr}</td>
            </tr>
        )
    })

    return (
        type == DataType.Gateway
        ?
        <>
        <h1>Gateway {liveDataInfo.uniqueId}</h1>
        <table class="table">
            <thead>
                <tr>
                <th scope="col">#</th>
                <th scope="col">Time</th>
                <th scope="col">Type</th>
                <th scope="col">RSSI</th>
                <th scope="col">SNR</th>
                </tr>
            </thead>
            <tbody>{ liveData }</tbody>
        </table>
        </>
        :
        <></>
    )
}
