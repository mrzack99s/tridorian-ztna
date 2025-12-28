import { useState, useEffect } from 'react';
import './App.css';
import { Greet, GetGateways, Connect, Disconnect, SignOut } from "../wailsjs/go/main/App";
import { WindowMinimise, Quit, EventsOn } from "../wailsjs/runtime/runtime";

interface Gateway {
    name: string;
    address: string;
    ping: string;
}

function App() {
    const [isConnected, setIsConnected] = useState(false);
    const [statusText, setStatusText] = useState("Ready");
    const [duration] = useState("00:00:00");
    const [domain, setDomain] = useState(localStorage.getItem("savedDomain") || "");
    const [showInput, setShowInput] = useState(true);
    const [gateways, setGateways] = useState<Gateway[]>([]);
    const [ip, setIp] = useState("");
    const [selectedGateway, setSelectedGateway] = useState<string>("");

    useEffect(() => {
        EventsOn("vpn_status", (status: string) => {
            console.log("VPN Status:", status);
            if (status.startsWith("Connected")) {
                setIsConnected(true);
                // Extract IP if present in the status message (Connected: {...json...})
                if (status.startsWith("Connected: ")) {
                    try {
                        const jsonStr = status.substring("Connected: ".length);
                        const data = JSON.parse(jsonStr);
                        if (data && data.assigned_ip) {
                            setIp(data.assigned_ip);
                        }
                    } catch (e) {
                        console.error("Failed to parse connection details", e);
                    }
                }
                setStatusText("Connected");
            } else if (status === "Disconnected") {
                setIsConnected(false);
                setStatusText("Disconnected");
                setIp("");
            } else {
                // Connecting, Authenticated, or Error
                setStatusText(status);
                if (status.startsWith("Error") || status.startsWith("Crash") || status.startsWith("Failed")) {
                    setIsConnected(false);
                }
            }
        });
    }, []);

    const handleLogin = () => {
        if (!domain) return;
        localStorage.setItem("savedDomain", domain);
        setStatusText("Authenticating...");

        Greet(domain).then((result) => {
            if (result.startsWith("Logged in as")) {
                setShowInput(false);
                fetchGateways(domain); // Pass domain
                setStatusText("Select Gateway");
            } else {
                setStatusText(result);
            }
        });
    }

    const fetchGateways = (domainName: string) => {
        GetGateways(domainName).then((result: any) => {
            // Wails returns map[string]string -> convert to Gateway[]
            // But backend returns []map[string]string, so it should be fine.
            // Need to ensure types match.
            setGateways(result || []);
            if (result && result.length > 0) {
                setSelectedGateway(result[0].address);
            }
        });
    }

    const toggleConnection = () => {
        if (isConnected) {
            setIsConnected(false);
            setStatusText("Disconnected");
            Disconnect().then(result => {
                setStatusText(result);
                // User remains logged in
            });
        } else {
            if (!selectedGateway) {
                setStatusText("Select a Gateway");
                return;
            }
            setStatusText("Connecting...");
            Connect(selectedGateway);
        }
    };

    const handleSignOut = () => {
        SignOut().then(() => {
            setIsConnected(false);
            setStatusText("Ready");
            setShowInput(true);
            setGateways([]);
        });
    };

    return (
        <div className="app-container">
            <header className="header" style={{ "--wails-draggable": "drag" } as any}>
                <div className="brand">
                    <div className="brand-icon">
                        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                            <path d="M12 2L2 7L12 12L22 7L12 2Z" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
                            <path d="M2 17L12 22L22 17" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
                            <path d="M2 12L12 17L22 12" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
                        </svg>
                    </div>
                    <span>Tridorian ZTNA</span>
                </div>

                <div className="header-right">
                    {!showInput && (
                        <button className="signout-btn" onClick={handleSignOut} title="Sign Out">
                            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                                <path d="M9 21H5C4.46957 21 3.96086 20.7893 3.58579 20.4142C3.21071 20.0391 3 19.5304 3 19V5C3 4.46957 3.21071 3.96086 3.58579 3.58579C3.96086 3.21071 4.46957 3 5 3H9" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
                                <path d="M16 17L21 12L16 7" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
                                <path d="M21 12H9" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
                            </svg>
                        </button>
                    )}
                    <div className={`status-badge ${isConnected ? '' : 'disconnected'}`}>
                        {isConnected ? 'SECURE' : 'OFFLINE'}
                    </div>

                    <div className="window-controls">
                        <button className="control-btn minimize" onClick={WindowMinimise}>
                            <svg width="10" height="1" viewBox="0 0 10 1" fill="none" xmlns="http://www.w3.org/2000/svg">
                                <rect width="10" height="1" fill="currentColor" />
                            </svg>
                        </button>
                        <button className="control-btn close" onClick={Quit}>
                            <svg width="10" height="10" viewBox="0 0 10 10" fill="none" xmlns="http://www.w3.org/2000/svg">
                                <path d="M1 1L9 9M9 1L1 9" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" />
                            </svg>
                        </button>
                    </div>
                </div>
            </header>

            <main className="main-content">
                <div className="connection-card">
                    {showInput ? (
                        <div className="login-form">
                            <h2 className="login-title">Sign In</h2>
                            <p className="login-subtitle">Enter your domain</p>

                            <div className="input-group">
                                <input
                                    type="text"
                                    className="domain-input"
                                    placeholder="acme.tridorian.com"
                                    value={domain}
                                    onChange={(e) => setDomain(e.target.value)}
                                />
                            </div>

                            <button className="primary-btn" onClick={handleLogin}>
                                Continue
                            </button>
                        </div>
                    ) : (
                        <>
                            {!isConnected && (
                                <div className="gateway-selector">
                                    <label className="selector-label">Data Center</label>
                                    <select
                                        className="gateway-select"
                                        value={selectedGateway}
                                        onChange={(e) => setSelectedGateway(e.target.value)}
                                    >
                                        {gateways.map((gw, idx) => (
                                            <option key={idx} value={gw.address}>
                                                {gw.name} ({gw.address})
                                            </option>
                                        ))}
                                    </select>
                                </div>
                            )}

                            <div
                                className={`big-switch ${isConnected ? 'active' : ''}`}
                                onClick={toggleConnection}
                            >
                                <svg className="switch-icon" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                                    <path d="M18.36 6.64a9 9 0 1 1-12.73 0" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
                                    <line x1="12" y1="2" x2="12" y2="12" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" />
                                </svg>
                            </div>

                            <div className="connection-status-text">
                                {statusText}
                            </div>

                            {isConnected && (
                                <div className="connection-details">
                                    <p>Duration: {duration}</p>
                                    <p>Connected to: {gateways.find(g => g.address === selectedGateway)?.name}</p>
                                    <p>IP: {ip}</p>
                                </div>
                            )}

                            {!isConnected && (
                                <div className="connection-details">
                                    <p className="workspace-domain">{domain}</p>
                                    <p>Click to secure your connection</p>
                                </div>
                            )}
                        </>
                    )}
                </div>
            </main>

            <footer className="footer">
                v1.0.0 • Tridorian ZTNA Client
            </footer>
        </div>
    )
}

export default App
