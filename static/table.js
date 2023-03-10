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

class FrameTable extends React.Component {
    state = {
        data: [],
        showOptions: [],
    }

    componentDidMount() {
        axios.get(this.props.path).then(res => {
            this.setState({ data: res.data })
        })
        this.props.setOptions(this.state.showOptions)
    }

    render() {
        return null
    }
}

const ByteBlocks = ({ size, value }) => {
    let hex = 
        typeof(value) === 'string' ? base64ToHex(value) :
        typeof(value) === 'number' ? numberToHex(BigInt(value), size) : null;

    return (
        <span 
            class="badge" 
            style={{
                color:"var(--bs-body-bg)", 
                backgroundColor:"var(--bs-light)"}}
        >{hex}</span>
    )
}

const EndDeviceTable = ({ path, showOptions, setOptions }) => {
    const [defaultOptions] = React.useState(["Appkey", "DevEui", "DevAddr"])
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
        return devEui & 0xFFFFFF
    }

    const endevices  = data.map((device, index) => { return (
        <tr>
            <th scope="row">{index + 1}</th>
            <td>{formatBusId(device.Id)}</td>
            {showOptions['Appkey'] ? <td><ByteBlocks value={device.Appkey} size={16}/></td> : null}
            {showOptions['DevEui'] ? <td><ByteBlocks value={device.DevEui} size={8}/></td> : null}
            {showOptions['DevAddr'] ? <td><ByteBlocks value={device.DevAddr} size={4}/></td> : null}
            <td><button className="btn btn-link p-1"><ThreeVerticalDots/></button></td>
            <td><DeleteButton deviceId={device.Id} path={path} doRefresh={setRefreshData}/></td>
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
                {showOptions['DevAddr'] ? <th scope="col">DevAddr</th> : null}
                <th scope="col">Actions</th>
                <th scope="col"></th>
                </tr>
            </thead>
            <tbody>{ endevices }</tbody>
        </table>
    )
}

const GatewayTable = ({ path, showOptions, setOptions }) => {

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
            {showOptions['Created Date'] ? <td>{gateway.Created}</td> : null}
            <td><button className="btn btn-link p-1"><ThreeVerticalDots/></button></td>
            <td><DeleteButton deviceId={gateway.Id} path={path} doRefresh={setRefreshData}/></td>
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
                <th scope="col">Actions</th>
                <th scope="col"></th>
                </tr>
            </thead>
            <tbody>{ gateways }</tbody>
        </table>
    )
}