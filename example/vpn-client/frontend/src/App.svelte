<script>
  import {StartLoginProcess, ConnectVPN, DisconnectVPN} from '../wailsjs/go/main/App.js'
  import {EventsOn} from '../wailsjs/runtime/runtime.js'

  let status = "Disconnected";
  let token = "";
  let serverIP = "localhost:6500";

  async function login() {
    status = "Logging in...";
    token = await StartLoginProcess();
    if(token) status = "Token Received. Ready.";
  }

  async function connect() {
    if(!token) return alert("Login first!");
    status = "Connecting...";
    await ConnectVPN(token, serverIP);
  }

  EventsOn("vpn-status", (msg) => status = msg);
  EventsOn("vpn-error", (err) => alert(err));
</script>

<main style="text-align:center; padding:50px; font-family:sans-serif;">
  <h1>My Corporate VPN</h1>
  <div style="background:#eee; padding:20px; margin:20px; border-radius:8px;">
    <b>Status:</b> {status}
  </div>
  
  <div style="display:flex; gap:10px; justify-content:center;">
    {#if !token}
      <button on:click={login} style="padding:10px; background:blue; color:white;">Login Google</button>
    {:else}
      <input bind:value={serverIP} placeholder="Server IP:Port" style="padding:10px;">
      <button on:click={connect} style="padding:10px; background:green; color:white;">Connect</button>
      <button on:click={DisconnectVPN} style="padding:10px; background:red; color:white;">Stop</button>
    {/if}
  </div>
</main>