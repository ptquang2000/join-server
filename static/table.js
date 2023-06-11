const DeleteButton = ({deviceId, path, doRefresh}) => {
    const [responseError, setResponseError] = React.useState({ show: false })
    const [show, setShow] = React.useState(false);
    const target = React.useRef(null)

    const onClicked = async () => {
        await axios.delete(path + "/" + deviceId).then(function (response) {
            console.log(response.data)
            doRefresh(true)
            setResponseError({ show: !show })
        }).catch(function (error) {
            setShow(!show)
            setResponseError({
                show: !show,
                statusCode: error.response.status,
            })
        })
    }
    return (
        <>
        <button 
        ref={target} 
        className="btn btn-link p-1" 
        onClick={onClicked} 
        ><TrashFill/></button>
        
        <ReactBootstrap.Overlay 
        rootClose
        onHide={() => setShow(false)}
        target={target.current} 
        show={show} 
        placement="bottom">
        {({
        placement: _placement,
        arrowProps: _arrowProps,
        show: _show,
        popper: _popper,
        hasDoneInitialMeasure: _hasDoneInitialMeasure,
        ...props
        }) => (
        <div
            {...props}
            style={{
            position: 'relative',
            backgroundColor: 'rgba(255, 100, 100, 0.85)',
            padding: '2px 10px',
            color: 'white',
            borderRadius: 3,
            ...props.style,
            }}
        >
            Status code: {responseError.statusCode}
        </div>
        )}
        </ReactBootstrap.Overlay>
        </>
    )
}

const LiveDataButton = ({deviceId, uniqueId, path, setLiveDataInfo}) => {
    const onClicked = () => {
        let liveDataInfo = path + "/" + deviceId
        setLiveDataInfo({path: liveDataInfo, uniqueId: uniqueId})
    }

    return (
        <>
        <button 
        className="btn btn-link p-1" 
        onClick={onClicked} 
        ><ThreeVerticalDots/></button>
        </>
    )
}

const FrameTable = ({ path, showOptions, setOptions, refreshTable }) => {
    const [defaultOptions] = React.useState(["Type", "RSSI", "SNR", "Mic", "Gateway ID"])
    const [data, setData] = React.useState([])
    const [refreshData, setRefreshData] = React.useState(true)

    const getData = () => {
        axios.get(path).then(res => {
            setData(res.data)
        })
        setOptions(defaultOptions)
    }

    React.useEffect(() => {
        if (refreshData)
        {
            getData()
        }
        setRefreshData(false)

    }, [path, refreshData])

    React.useEffect(() => {
        getData()
    }, [refreshTable])

    const frames  = data.map((frame, index) => { return (
        <tr>
            <th scope="row">{index + 1}</th>
            {showOptions['Type'] ? <td>{GetFrameType(frame.FrameType)}</td> : null}
            {showOptions['RSSI'] ? <td>{frame.Rssi}</td> : null}
            {showOptions['SNR'] ? <td>{frame.Snr}</td> : null}
            {showOptions['Mic'] ? <th scope="col">{frame.Mic}</th> : null}
            {showOptions['Gateway ID'] ? <th scope="col">{frame.GatewayID}</th> : null}
        </tr>
    )})

    return (
        <table class="table">
            <thead>
                <tr>
                <th scope="col">#</th>
                {showOptions['Type'] ? <th scope="col">Type</th> : null}
                {showOptions['RSSI'] ? <th scope="col">RSSI</th> : null}
                {showOptions['SNR'] ? <th scope="col">SNR</th> : null}
                {showOptions['Mic'] ? <th scope="col">Mic</th> : null}
                {showOptions['Gateway ID'] ? <th scope="col">Gateway ID</th> : null}
                </tr>
            </thead>
            <tbody>{ frames }</tbody>
        </table>
    )
}

const ByteBlocks = ({ size, value }) => {
    let hex = 
        typeof(value) === 'string' ? base64ToHex(value) :
        typeof(value) === 'bigint' ? numberToHex(value, size) : null;

    return (
        <span 
            class="badge" 
            style={{
                color:"var(--bs-body-bg)", 
                backgroundColor:"var(--bs-light)"}}
        >{hex}</span>
    )
}

const EndDeviceTable = ({ path, showOptions, setOptions, setLiveDataInfo }) => {
    const [defaultOptions] = React.useState(["Appkey", "DevEui", "JoinEui", "DevAddr"])
    const [data, setData] = React.useState([])
    const [refreshData, setRefreshData] = React.useState(true)

    React.useEffect(() => {
        if (refreshData)
        {
            axios.get(path).then(res => {
                setData(res.data)
            })
            setOptions(defaultOptions)
        }
        setRefreshData(false)
    }, [path, refreshData])

    const formatBusId = (devEui) => {
        return (BigInt(devEui) & BigInt(0xFFFFFF)).toString()
    }

    const endDevices = data.map((device, index) => { return (
        <tr>
            <th scope="row">{index + 1}</th>
            <td>{formatBusId(device.DevEui)}</td>
            {showOptions['Appkey'] ? <td><ByteBlocks value={device.Appkey} size={16}/></td> : null}
            {showOptions['DevEui'] ? <td><ByteBlocks value={BigInt(device.DevEui)} size={8}/></td> : null}
            {showOptions['JoinEui'] ? <td><ByteBlocks value={BigInt(device.JoinEui)} size={8}/></td> : null}
            {showOptions['DevAddr'] ? <td><ByteBlocks value={BigInt(device.DevAddr)} size={4}/></td> : null}
            <td><LiveDataButton deviceId={device.ID} uniqueId={BigInt(device.DevEui)} path={path} setLiveDataInfo={setLiveDataInfo}/></td>
            <td><DeleteButton deviceId={device.ID} path={path} doRefresh={setRefreshData}/></td>
        </tr>
    )})

    return (
        <table class="table">
            <thead>
                <tr>
                <th scope="col">#</th>
                <th scope="col">Bus ID</th>
                {showOptions['Appkey'] ? <th scope="col">AppKey</th> : null}
                {showOptions['DevEui'] ? <th scope="col">DevEUI</th> : null}
                {showOptions['JoinEui'] ? <th scope="col">JoinEui</th> : null}
                {showOptions['DevAddr'] ? <th scope="col">DevAddr</th> : null}
                <th scope="col">Live Data</th>
                <th scope="col"></th>
                </tr>
            </thead>
            <tbody>{ endDevices }</tbody>
        </table>
    )
}

const GatewayTable = ({ path, showOptions, setOptions, setLiveDataInfo }) => {

    const [defaultOptions] = React.useState(["Admin", 'Created Date'])
    const [data, setData] = React.useState([])
    const [refreshData, setRefreshData] = React.useState(true)

    React.useEffect(() => {
        if (refreshData)
        {
            axios.get(path).then(res => {
                setData(res.data)
            })
            setOptions(defaultOptions)
        }
        setRefreshData(false)
    }, [path, refreshData])

    const gateways  = data.map((gateway, index) => { return (
        <tr>
            <th scope="row">{index + 1}</th>
            <td>{gateway.Username}</td>
            {showOptions['Admin'] ? 
            <td style={{color: gateway.Is_superuser ? "var(--bs-success)" : "var(--bs-danger)"}}>
                {gateway.Is_superuser ? <CheckIcon/> : <CrossIcon/>}
            </td> 
            : null}
            {showOptions['Created Date'] ? <td>{gateway.CreatedAt}</td> : null}
            <td><LiveDataButton deviceId={gateway.ID} uniqueId={gateway.ID} path={path} setLiveDataInfo={setLiveDataInfo}/></td>
            <td><DeleteButton deviceId={gateway.ID} path={path} doRefresh={setRefreshData}/></td>
        </tr>
    )})

    return (
        <table class="table">
            <thead>
                <tr>
                <th scope="col">#</th>
                <th scope="col">Station</th>
                {showOptions['Admin'] ? <th scope="col">Admin</th> : null}
                {showOptions['Created Date'] ? <th scope="col">Created Date</th> : null}
                <th scope="col">Live Data</th>
                <th scope="col"></th>
                </tr>
            </thead>
            <tbody>{ gateways }</tbody>
        </table>
    )
}
