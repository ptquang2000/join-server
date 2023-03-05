const ContainerHeader = ({ options, OnInputChanged }) => {
    const onSearch = (e) => {
        e.preventDefault()
    }
    const [show, setShow] = React.useState(false);
    const target = React.useRef(null)
    const [currentOptions, setCurrentOptions] = React.useState({})

    const optionList = options.map(option => 
            <ReactBootstrap.ListGroup.Item className="input-group-prepend">
                <input 
                    type="checkbox" 
                    id={option} 
                    defaultChecked={true}
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
        if (options.length != 0)
        {
            setCurrentOptions(options.reduce((obj, key) => ({ ...obj, [key]: true}), {}))
        }
    }, [options])

    React.useEffect(() => {
        OnInputChanged(currentOptions)
    }, [currentOptions])

    return (
        <div className="text-start">
            <h1>Container Name</h1>
            <p>Container Description</p>

            <div className="d-flex justify-content-end mt-5">
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
                <ReactBootstrap.Overlay target={target.current} show={show} placement="top">
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
                    <ReactBootstrap.ListGroup>{optionList}</ReactBootstrap.ListGroup>
                    </div>
                    )}
                </ReactBootstrap.Overlay>
            </div>
        </div>
    )
}

const Containers = ({ path, type }) => {
    const [options, setOptions] = React.useState([])
    const [showOptions, setShowOptions] = React.useState({})

    const table = 
        type == DataType.Gateway ? <GatewayTable path={path} setOptions={setOptions} showOptions={showOptions}/> :
        type == DataType.EndDevices ? <EndDeviceTable path={path} setOptions={setOptions} showOptions={showOptions}/> :
        type == DataType.Frames ? <FrameTable path={path} setOptions={setOptions} showOptions={showOptions}/> : null
    return (
        <main className="container-fluid text-center">
            <ContainerHeader 
                options={options} 
                OnInputChanged={setShowOptions}
            />
            {table}
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
            <div className="container-lg position-absolute top-50 start-50 translate-middle">
                <div className="container-lg d-flex flex-row">
                    <Navbars onTabChanged={this.onTabChanged}/>
                    <Containers path={this.state.path} type={this.state.dataType}/>
                </div>
            </div>
        )
    }
}

ReactDOM.render(<App/>, document.getElementById('app'))