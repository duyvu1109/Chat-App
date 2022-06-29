// import React from "react";
// import { useState } from "react";
import React from "react";
import { ChatUI } from "./components/chat/chat";

var client = new WebSocket("ws://localhost:8080/ws");
client.binaryType = "arraybuffer";

function App() {
  return (
    <React.Fragment>
      {/* <MyForm client={client} /> */}
      <br/>
      <ChatUI client= {client}/>
    </React.Fragment>
  );
}

export default App;
