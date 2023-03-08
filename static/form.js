const ErrorToast = ({ show, onClosed, statusCode, message }) => {
    return (
        <ReactBootstrap.Toast show={show} onClose={onClosed}>
            <ReactBootstrap.Toast.Header className="d-flex justify-content-between">
            <strong>Status: {statusCode}</strong>
            </ReactBootstrap.Toast.Header>
            <ReactBootstrap.Toast.Body>{message}</ReactBootstrap.Toast.Body>
        </ReactBootstrap.Toast>
    )
}

const GatewayForm = ({ path, setSuccess }) => {
    const [validated, setValidated] = React.useState(false)
    const [errors, setErrors] = React.useState({username: false, password: false, salt: false})
    const [responseError, setResponseError] = React.useState({ show: false })

    const onSubmitted = async (e) => {
        e.preventDefault()

        setErrors({
            username: e.target[0].value == '', 
            password: e.target[1].value == '', 
            salt: e.target[2].value == '',
        })
        
        if (e.target[0].value != '' &&
            e.target[1].value != '' &&
            e.target[2].value != '')
        {
            setValidated(true)
            await axios.post(path, {
                Username: e.target[0].value,
                Password: e.target[1].value,
                Salt: e.target[2].value,
                Is_superuser: e.target[3].checked
            }).then(function (response) {
                console.log(response.data)
                setResponseError({ show: false })
                setSuccess(true)
            }).catch(function (error) {
                setResponseError({
                    show: true,
                    statusCode: error.response.status,
                    message: error.response.data.Message,
                })
            })
        }
    }

    return (
        <ReactBootstrap.Row>
            <ReactBootstrap.Col xs={1}></ReactBootstrap.Col>
            <ReactBootstrap.Col xs={5} className="fw-bold">
                
                <ReactBootstrap.Form noValidate validated={validated} onSubmit={onSubmitted}>
                    <ReactBootstrap.Form.Group controlId="Username">
                        <ReactBootstrap.Form.Label class>Mqtt Username</ReactBootstrap.Form.Label>
                        <ReactBootstrap.Form.Control 
                        type="text" 
                        placeholder="Enter Username" 
                        required
                        isInvalid={errors.usename}
                        />
                        <ReactBootstrap.Form.Control.Feedback type="invalid">
                            Could not be null
                        </ReactBootstrap.Form.Control.Feedback>
                    </ReactBootstrap.Form.Group>
            
                    <ReactBootstrap.Form.Group controlId="Password">
                        <ReactBootstrap.Form.Label class>Mqtt Password</ReactBootstrap.Form.Label>
                        <ReactBootstrap.Form.Control 
                        type="password" 
                        placeholder="Enter Password" 
                        required
                        isInvalid={errors.password}
                        />
                        <ReactBootstrap.Form.Control.Feedback type="invalid">
                            Could not be null
                        </ReactBootstrap.Form.Control.Feedback>
                    </ReactBootstrap.Form.Group>
            
                    <ReactBootstrap.Form.Group controlId="Salt">
                        <ReactBootstrap.Form.Label class>Password Salt</ReactBootstrap.Form.Label>
                        <ReactBootstrap.Form.Control 
                        type="text" 
                        placeholder="Enter Salt"
                        required    
                        isInvalid={errors.salt}
                        />
                        <ReactBootstrap.Form.Control.Feedback type="invalid">
                            Could not be null
                        </ReactBootstrap.Form.Control.Feedback>
                    </ReactBootstrap.Form.Group>
            
                    <ReactBootstrap.Form.Group controlId="IsSuperUser">
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
            <ReactBootstrap.Col xs={5} className="d-flex align-items-center">
                <ErrorToast 
                {...responseError}
                onClosed={() => setResponseError({show:false})}
                />
            </ReactBootstrap.Col>
        </ReactBootstrap.Row>
    )
}

const EndDeviceForm = ({ path, setSuccess }) => {
    const [errors, setErrors] = React.useState({devEui: false, appKey: false})
    const [responseError, setResponseError] = React.useState({ show: false })

    const initialDevEui = ['f', 'e', 'f', 'f', 'f', 'f', '0','0', '0', '0', '0']
    const [devEui, setDevEui] = React.useState(initialDevEui)
    const devEuiRef = React.useRef(null)
    const onDevEuiChanged = (e) => {
        const input = e.target.value;
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
        if (e.key == 'Backspace' && devEui.length > initialDevEui.length)
        {
            devEui.pop()
            setDevEui([...devEui])
        }
    }
    React.useEffect(() => {
        devEuiRef.current.value = generateHtmlEntity(16, 8226, devEui)
    }, [devEui])

    const [appKey, setAppKey] = React.useState('')
    const onAppKeyChanged = () => {
        axios.get("/appkey").then(res => {
            setAppKey(res.data)
        })
    }
    const appKeyRef = React.useRef(null)
    React.useEffect(() => {
        var hexArr = base64ToHexArray(appKey)
        appKeyRef.current.value = generateHtmlEntity(hexArr.length, null, hexArr)
    }, [appKey])

    const [validated, setValidated] = React.useState(false)
    const onSubmitted = async (e) => {
        e.preventDefault()

        setErrors({
            devEui: devEui.length != 16, 
            appKey: appKey.length == '',
        })

        if (devEui.length == 16 && appKey.length != '')
        {
            setValidated(true)
            await axios.post(path, {
                NetId: 0,
                JoinEui: 0,
                DevEui: hexToNum(devEui),
                AppKey: appKey,
            }).then(function (response) {
                console.log(response.data)
                setResponseError({ show: false })
                setSuccess(true)
            }).catch(function (error) {
                setResponseError({
                    show: true,
                    statusCode: error.response.status,
                    message: error.response.data.Message,
                })
            });
        }
    }
    return (
        <ReactBootstrap.Row>
        <ReactBootstrap.Col xs="1"/>
        <ReactBootstrap.Col className="fw-bold">
            
            <ReactBootstrap.Form noValidate validated={validated} onSubmit={onSubmitted}>
                <ReactBootstrap.Form.Group controlId="formNetId">
                    <ReactBootstrap.Form.Label>NetID</ReactBootstrap.Form.Label>
                    <ReactBootstrap.Form.Control 
                    style={{width: "110px"}}
                    type="text"
                    size="sm"
                    placeholder={generateHtmlEntity(6, 48)}
                    />
                </ReactBootstrap.Form.Group>
        
                <ReactBootstrap.Form.Group controlId="formJoinEui">
                    <ReactBootstrap.Form.Label>JoinEUI</ReactBootstrap.Form.Label>
                    <ReactBootstrap.Form.Control 
                    style={{width: "238px"}}
                    type="text" 
                    size="sm"
                    placeholder={generateHtmlEntity(16, 48)}
                    />
                </ReactBootstrap.Form.Group>
        
                <ReactBootstrap.Form.Group controlId="formDevEui">
                    <ReactBootstrap.Form.Label>DevEUI</ReactBootstrap.Form.Label>
                    <ReactBootstrap.Form.Control 
                    ref={devEuiRef}
                    style={{width: "238px"}}
                    type="text" 
                    size="sm"
                    onChange={onDevEuiChanged}
                    onKeyDown={onDevEuiKeyPressed}
                    isInvalid={errors.devEui}
                    required
                    />
                    <ReactBootstrap.Form.Control.Feedback type="invalid">
                        DevEUI has to be 16 bytes
                    </ReactBootstrap.Form.Control.Feedback>
                </ReactBootstrap.Form.Group>
                
                <ReactBootstrap.Form.Group as={ReactBootstrap.Row} controlId="formAppKey">
                    <ReactBootstrap.Form.Label>AppKey</ReactBootstrap.Form.Label>
                    <ReactBootstrap.Col>
                        <ReactBootstrap.Form.Control 
                        ref={appKeyRef}
                        style={{width: "440px"}}
                        type="text" 
                        size="sm"
                        placeholder={generateHtmlEntity(32, 8226)} 
                        isInvalid={errors.appKey}
                        required
                        />
                        <ReactBootstrap.Form.Control.Feedback type="invalid">
                            Please generate Appkey
                        </ReactBootstrap.Form.Control.Feedback>
                    </ReactBootstrap.Col>
                    <ReactBootstrap.Col>
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
        <ReactBootstrap.Col  xs="5" className="d-flex align-items-center">
            <ErrorToast 
                {...responseError}
                onClosed={() => setResponseError({show:false})}
            />
        </ReactBootstrap.Col>
    </ReactBootstrap.Row>
    )
}