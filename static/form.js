const GatewayForm = ({}) => {
    return (
        <ReactBootstrap.Row>
            <ReactBootstrap.Col xs={1}></ReactBootstrap.Col>
            <ReactBootstrap.Col xs={5} className="fw-bold">
                
                <ReactBootstrap.Form>
                    <ReactBootstrap.Form.Group controlId="formUsername">
                        <ReactBootstrap.Form.Label class>Mqtt Username</ReactBootstrap.Form.Label>
                        <ReactBootstrap.Form.Control type="email" placeholder="Enter Username" />
                    </ReactBootstrap.Form.Group>
            
                    <ReactBootstrap.Form.Group controlId="formPassword">
                        <ReactBootstrap.Form.Label class>Mqtt Password</ReactBootstrap.Form.Label>
                        <ReactBootstrap.Form.Control type="password" placeholder="Enter Password" />
                    </ReactBootstrap.Form.Group>
            
                    <ReactBootstrap.Form.Group controlId="formSalt">
                        <ReactBootstrap.Form.Label class>Password Salt</ReactBootstrap.Form.Label>
                        <ReactBootstrap.Form.Control type="text" placeholder="Enter Salt" />
                    </ReactBootstrap.Form.Group>
            
                    <ReactBootstrap.Form.Group controlId="formIsSuperUser">
                        <ReactBootstrap.Form.Check type="checkbox" label="Is Super User?" />
                    </ReactBootstrap.Form.Group>

                    <ReactBootstrap.Button 
                    variant="primary" 
                    type="submit"
                    style={{
                        backgroundColor:"var(--bs-primary-text)"
                    }}
                    className="mt-3"
                    >
                        Register Gateway
                    </ReactBootstrap.Button>

                </ReactBootstrap.Form>
            
            </ReactBootstrap.Col>
            <ReactBootstrap.Col xs={5}></ReactBootstrap.Col>
        </ReactBootstrap.Row>
    )
}

const EndDeviceForm = ({}) => {
    const initialDevEui = ['F', 'E', 'F', 'F', 'F', 'F', '0','0', '0', '0', '0']
    const [devEui, setDevEui] = React.useState(initialDevEui)
    const devEuiRef = React.useRef(null)
    const onDevEuiChanged = (e) => {
        const input = e.target.value;
        console.log(input[input.length - 1])
        if (/^[0-9a-f]+$/.test(input[input.length - 1]) || input === "") {
            if (devEuiRef.current.value[devEuiRef.current.value.length - 2] == String.fromCharCode(8226))
            {
                devEui.push(input[input.length - 1])
                setDevEui([...devEui])
            }
        }
        devEuiRef.current.value = generateHtmlEntity(16, 8226, devEui)
    }
    const onDevEuiKeyPressed = (e) => {
        if (e.key == 'Backspace')
        {
            if (devEui.length > initialDevEui.length)
            {
                devEui.pop()
                setDevEui([...devEui])
            }
        }
    }
    React.useEffect(() => {
        devEuiRef.current.value = generateHtmlEntity(16, 8226, devEui)
        console.log(devEui.length)
    }, [devEui])


    const [appKey, setAppKey] = React.useState([])
    const onAppKeyChanged = (e) => {
        axios.get("/appkey").then(res => {
            setAppKey(base64ToHexArray(res.data))
        })
    }
    const appKeyRef = React.useRef(null)
    React.useEffect(() => {
        appKeyRef.current.value = generateHtmlEntity(appKey.length, null, appKey)
    }, [appKey])

    return (
        <ReactBootstrap.Row>
        <ReactBootstrap.Col xs="1"/>
        <ReactBootstrap.Col className="fw-bold">
            
            <ReactBootstrap.Form>
                <ReactBootstrap.Form.Group controlId="formNetId">
                    <ReactBootstrap.Form.Label class>NetID</ReactBootstrap.Form.Label>
                    <ReactBootstrap.Form.Control 
                    style={{width: "84px"}}
                    type="text"
                    size="sm"
                    placeholder={generateHtmlEntity(6, 48)}
                    readOnly
                    />
                </ReactBootstrap.Form.Group>
        
                <ReactBootstrap.Form.Group controlId="formJoinEui">
                    <ReactBootstrap.Form.Label class>JoinEUI</ReactBootstrap.Form.Label>
                    <ReactBootstrap.Form.Control 
                    style={{width: "212px"}}
                    type="text" 
                    size="sm"
                    placeholder={generateHtmlEntity(16, 48)}
                    readOnly
                    />
                </ReactBootstrap.Form.Group>
        
                <ReactBootstrap.Form.Group controlId="formDevEui">
                    <ReactBootstrap.Form.Label class>DevEUI</ReactBootstrap.Form.Label>
                    <ReactBootstrap.Form.Control 
                    ref={devEuiRef}
                    style={{width: "212px"}}
                    type="text" 
                    size="sm"
                    onChange={onDevEuiChanged}
                    onKeyDown={onDevEuiKeyPressed}
                    />
                </ReactBootstrap.Form.Group>
                
                <ReactBootstrap.Form.Group as={ReactBootstrap.Row} controlId="formAppKey">
                    <ReactBootstrap.Form.Label class>AppKey</ReactBootstrap.Form.Label>
                    <ReactBootstrap.Col sm="8">
                        <ReactBootstrap.Form.Control 
                        ref={appKeyRef}
                        style={{width: "414px"}}
                        type="text" 
                        size="sm"
                        placeholder={generateHtmlEntity(32, 8226)} 
                        />
                    </ReactBootstrap.Col>
                    <ReactBootstrap.Col sm="1">
                        <ReactBootstrap.Button
                        variant="outline-light" 
                        type="button"
                        onClick={onAppKeyChanged}
                        >
                            <ArrowCounterClockwise/>
                        </ReactBootstrap.Button>
                    </ReactBootstrap.Col>

                </ReactBootstrap.Form.Group>

                <ReactBootstrap.Button 
                variant="primary" 
                type="submit"
                style={{
                    backgroundColor:"var(--bs-primary-text)"
                }}
                className="mt-3"
                >
                    Register Device
                </ReactBootstrap.Button>

            </ReactBootstrap.Form>
        
        </ReactBootstrap.Col>
    </ReactBootstrap.Row>
    )
}