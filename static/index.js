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
                <form className="d-flex m-0" role="search" onSubmit={onSearch}>
                    <input className="form-control align-self-center" type="search" placeholder="Search" aria-label="Search"/>
                </form>
                <button
                    ref={target}
                    type="button" 
                    className="btn btn-link p-1" 
                    onClick={() => setShow(!show)}
                >
                    <ThreeVerticalDots/>
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
    const [page, setPage] = React.useState(true)

    const [options, setOptions] = React.useState([])
    const [showOptions, setShowOptions] = React.useState({})


    const table = 
        type == DataType.Gateway ? <GatewayTable path={path} setOptions={setOptions} showOptions={showOptions}/> :
        type == DataType.EndDevices ? <EndDeviceTable path={path} setOptions={setOptions} showOptions={showOptions}/> :
        type == DataType.Frames ? <FrameTable path={path} setOptions={setOptions} showOptions={showOptions}/> : 
        null

    const [successReq, setSuccessReq] = React.useState(false)
    const form = 
        type == DataType.Gateway ? <GatewayForm path={path} setSuccess={setSuccessReq}/> :
        type == DataType.EndDevices ? <EndDeviceForm path={path} setSuccess={setSuccessReq}/> :
        null
    React.useEffect(() => {
        if (successReq && !page)
        {
            setPage(true)
        }
    }, [successReq])

    return (
        <main className="container-fluid p-3" style={{backgroundColor:"var(--bs-info-border-subtle)"}}>
            {
                page == true
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
                    null
                }
                </>
                :
                <>
                <ReactBootstrap.Button 
                variant="link"
                className="ps-0"
                onClick={(_) => {setPage(!page)}}
                >
                    <ChevronLeft/>
                </ReactBootstrap.Button>
                {form}
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
            <div className="container-lg position-absolute top-50 start-50 translate-middle border border-light-subtle rounded p-0">
                <div className="container-lg d-flex flex-row p-0">
                    <Navbars onTabChanged={this.onTabChanged}/>
                    <Containers path={this.state.path} type={this.state.dataType}/>
                </div>
            </div>
        )
    }
}

ReactDOM.render(<App/>, document.getElementById('app'))