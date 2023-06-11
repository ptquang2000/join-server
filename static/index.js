const ContainerHeader = ({ type, options, OnInputChanged }) => {
    const onSearch = (e) => {
        e.preventDefault()
    }
    const [show, setShow] = React.useState(false);
    const target = React.useRef(null)
    const [currentOptions, setCurrentOptions] = React.useState({})

    const optionList = options.map(option => 
            <ReactBootstrap.ListGroup.Item className="input-group-prepend border-0">
                <input 
                    type="checkbox" 
                    id={option} 
                    checked={currentOptions[option]}
                    onChange={(e) => {
                        const newOptions = {...currentOptions}
                        newOptions[option] = e.target.checked
                        setCurrentOptions(newOptions)
                    }}
                />
                <span class="ms-1">{option}</span>
            </ReactBootstrap.ListGroup.Item>
    )
    React.useEffect(() => {
        setCurrentOptions(options.reduce((obj, key) => ({ ...obj, [key]: true}), {}))
    }, [options])

    React.useEffect(() => {
        OnInputChanged(currentOptions)
    }, [currentOptions])

    const description = 
        type == DataType.Gateway ? "Table of Gateways" :
        type == DataType.EndDevices ? "Table of End Devices" :
        type == DataType.Frames ? "Table of Frames" : null

    return (
        <div className="text-start">
            <h1>{type}</h1>
            <p>{description}</p>

            <div className="d-flex justify-content-end">
                {/* <form className="d-flex m-0" role="search" onSubmit={onSearch}>
                    <input className="form-control align-self-center" type="search" placeholder="Search" aria-label="Search"/>
                </form> */}
                <button
                    ref={target}
                    type="button" 
                    className="btn btn-link p-1" 
                    onClick={() => setShow(!show)}
                >
                    Options
                    {/* <ThreeVerticalDots/> */}
                </button>
                <ReactBootstrap.Overlay 
                target={target.current} 
                rootClose 
                show={show} 
                placement="bottom"
                onHide={() => setShow(false)}
                >
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
                        position: 'absolute',
                        padding: '2px 10px',
                        color: 'white',
                        borderRadius: 3,
                        ...props.style,
                        }}
                    >
                    <ReactBootstrap.ListGroup className="border border-light-subtle">
                        {optionList}
                        <ReactBootstrap.ListGroup.Item className="border-0">
                            <ReactBootstrap.Button  
                                variant="link" 
                                className="text-decoration-none m-0 pe-2 ps-0 py-0"
                                onClick={(e) => {
                                    setCurrentOptions(options.reduce((obj, key) => ({ ...obj, [key]: false}), {}))
                                }}
                            >
                                Hide all
                            </ReactBootstrap.Button>
                            <ReactBootstrap.Button 
                                variant="link" 
                                className="text-decoration-none m-0 pe-0 ps-2 py-0"
                                onClick={(e) => {
                                    setCurrentOptions(options.reduce((obj, key) => ({ ...obj, [key]: true}), {}))
                                }}
                            >
                                Show all
                            </ReactBootstrap.Button>
                        </ReactBootstrap.ListGroup.Item>
                    </ReactBootstrap.ListGroup>
                    </div>
                    )}
                </ReactBootstrap.Overlay>
            </div>
        </div>
    )
}

const Containers = ({ path, type }) => {
    const TABLE = 0
    const FORM = 1
    const LIVEDATA = 2
    const [page, setPage] = React.useState(TABLE)
    const [refreshTable, setRefreshTable] = React.useState(false)
    const [liveDataInfo, setLiveDataInfo] = React.useState({})

    const [options, setOptions] = React.useState([])
    const [showOptions, setShowOptions] = React.useState({})
    React.useEffect(() => {
        setPage(TABLE)
    }, [path])
    
    const table = 
        type == DataType.Gateway ? <GatewayTable path={path} setOptions={setOptions} showOptions={showOptions} setLiveDataInfo={setLiveDataInfo}/> :
        type == DataType.EndDevices ? <EndDeviceTable path={path} setOptions={setOptions} showOptions={showOptions} setLiveDataInfo={setLiveDataInfo}/> :
        type == DataType.Frames ? <FrameTable path={path} setOptions={setOptions} showOptions={showOptions} refreshTable={refreshTable}/> : 
        null

    const [successReq, setSuccessReq] = React.useState(false)
    const form = 
        type == DataType.Gateway ? <GatewayForm path={path} setSuccess={setSuccessReq}/> :
        type == DataType.EndDevices ? <EndDeviceForm path={path} setSuccess={setSuccessReq}/> :
        null

    React.useEffect(() => {
        setPage(LIVEDATA)
    }, [liveDataInfo])
    const liveDataTable =
        JSON.stringify(liveDataInfo) === '{}' ? null :
        type == DataType.Gateway ? <GatewayLiveTable type={type} liveDataInfo={liveDataInfo}/> :
        type == DataType.EndDevices ? <EndDeviceLiveTable type={type} liveDataInfo={liveDataInfo}/> :
        null
    
    React.useEffect(() => {
        if (successReq && page == FORM)
        {
            setPage(TABLE)
            setSuccessReq(false)
        }
    }, [successReq])

    return (
        <main className="flex-fill p-3" style={{backgroundColor:"var(--bs-info-border-subtle)"}}>
            {
                page == TABLE
                ?
                <>
                <ContainerHeader 
                    type={type}
                    options={options} 
                    OnInputChanged={setShowOptions}
                />
                {table}
                {
                    type != DataType.Frames ?
                    <div  className="d-flex justify-content-end">
                        <ReactBootstrap.Button 
                        variant="link"
                        className="pe-2"
                        onClick={(_) => {setPage(!page)}}
                        >
                            <PlusCircleFill/>
                        </ReactBootstrap.Button>
                    </div>
                    :
                    <div  className="d-flex justify-content-end">
                        <ReactBootstrap.Button 
                        variant="link"
                        className="pe-2"
                        onClick={(_) => {setRefreshTable(!refreshTable)}}
                        >
                            <ArrowCounterClockwise/>
                        </ReactBootstrap.Button>
                    </div>
                }
                </>
                :
                page == FORM
                ?
                <>
                <ReactBootstrap.Button 
                variant="link"
                className="ps-0"
                onClick={(_) => {setPage(TABLE)}}
                >
                    <ChevronLeft/>
                </ReactBootstrap.Button>
                {form}
                </>
                :
                <>
                <ReactBootstrap.Button 
                variant="link"
                className="ps-0"
                onClick={(_) => {setPage(TABLE)}}
                >
                    <ChevronLeft/>
                </ReactBootstrap.Button>
                {liveDataTable}
                </>
            }
        </main>
    )
}

class App extends React.Component {
    state = {
        path : window.location.href,
        dataType: -1,
    }

    onTabChanged = (path, type) => {
        this.setState({ 
            path: path,
            dataType: type,
        })
    }

    render() {
        return (
            <div className="border border-light-subtle rounded p-0">
                <div className="vh-100 d-flex flex-row">
                    <Navbars onTabChanged={this.onTabChanged}/>
                    <Containers path={this.state.path} type={this.state.dataType}/>
                </div>
            </div>
        )
    }
}

ReactDOM.render(<App/>, document.getElementById('app'))
