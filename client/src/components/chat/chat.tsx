import "./chat.css";
import React, { useState } from "react";
import { useEffect } from "react";
import { api } from "../../api/pb";

export function ChatUI(props: { client: WebSocket }) {
  const [user, setuser] = useState<any | null>(null);
  const [room, setroom] = useState<any | null>(null);
  const [msg, setmsg] = useState("");
  const [messresponse, setMessresponse] = useState(""); // 
  // Message List
  const [messages, setMessages] = useState<string[]>([]); //read
  // Room List
  const [rooms, setRooms] = useState(Array<string>(5));
  // // Active users List
  // const [users, setUsers]= useState([]);


  // submit LoginRequest
  const handleSubmitRoom = (event: any) => {
    event.preventDefault();
    var msg = new api.CreateRoomRequest({
        user: user,
        room: room,
      });
    var message = new api.Message({ create_room_request: msg });
    props.client.send(message.serialize());
  };


  const handleSubmitMessage = (event: any) => {
    event.preventDefault();
    var message = new api.ChatRequest({
        user: user,
        msg: msg,
        room: room,
      });
      var messageSent = new api.Message( {chat_request: message} )
      props.client.send(messageSent.serialize());
  };

  // Recieve Message Replies /////////////////////////////////////////////

  useEffect(() => {
    props.client.onmessage = function (event) {
      var msg = event.data;
      var converted_msg = new Uint8Array(msg);
      console.log("onmessage:", event.data, typeof event.data);
      try {
        var pbMessage = api.Message.deserialize(converted_msg);
        console.log("Converted message: " + pbMessage);
        if (pbMessage.chat_reply) {
          const newResponse = pbMessage.chat_reply.msg;
          setMessresponse(newResponse);
          console.log("Message: ", newResponse);

          let message_list = messages
          message_list.push(pbMessage.chat_reply.user + ": " + pbMessage.chat_reply.msg);
          setMessages(message_list);

        } else if (pbMessage.create_room_reply) {
          // Update Room List
        //   let room_list = [];
        //   room_list = pbMessage.create_room_reply;
        //   setRooms(room_list);
        //   //
        //   const newResponse = pbMessage.join_room_reply;
        //   setMessresponse(newResponse);
          console.log("Create room");
        } 
      } catch (err) {
        console.log(err);
      }
    };
  }, []);

  return (
    <React.Fragment>
      <div className="container">
        <div className="row no-gutters">
          <div className="col-md-3 border-right">
            <div className="settings-tray">
              {/* <span className="settings-tray--right">
                <i className="material-icons">cached</i>
                <i className="material-icons">menu</i>
              </span> */}
            </div>

            {/* <div className="search-box">
              <div className="input-wrapper">
                <i className="material-icons">search</i>
                <input placeholder="Search here" type="text" />
              </div>
            </div> */}

            <form
              onSubmit={(e) => {
                handleSubmitRoom(e);
              }}
            >
              <div className="row">
                <div className="col-12">
                  <div className="chat-box-tray">
                    <label>user: </label>
                    <input
                      name="user"
                      value={user || ""}
                      type="text"
                      onChange={(e) => setuser(e.target.value)}
                      placeholder="Type your user here..."
                    />
                  </div>
                </div>
              </div>
            </form>
            <br />

            <form onSubmit={(e) => handleSubmitRoom(e)}>
              <div className="row">
                <div className="col-12">
                  <div className="chat-box-tray">
                    <label>room: </label>
                    <input
                      name="room"
                      value={room || ""}
                      type="text"
                      onChange={(e) => setroom(e.target.value)}
                      placeholder="Type your room here..."
                    />
                    <button type="submit" className="material-icons">
                      send
                    </button>
                  </div>
                </div>
              </div>
            </form>

            <div>
              {rooms.map((txt, index) => (
                <div key={index}className="friend-drawer friend-drawer--onhover">
                  <div className="text">
                    <h6>[Room] {txt}</h6>
                    <p className="text-muted">Hey, you're arrested!</p>
                  </div>
                  <span className="time text-muted small">13:21</span>
                </div>
              ))}
            </div>
          </div>

          <div className="col-md-9">
            <div className="settings-tray">
              <div className="friend-drawer no-gutters friend-drawer--grey">
                <div className="text">
                  {/* <h6>Robo Cop</h6> */}
                  {/* <p className="text-muted">
                    Layin' down the law since like before Christ...
                  </p> */}
                </div>
                {/* <span className="settings-tray--right">
                  <i className="material-icons">cached</i>
                  <i className="material-icons">menu</i>
                </span> */}
              </div> 
            </div>

            <div className="chat-panel">
              <div className="row no-gutters">
                <div className="col-md-3">
                  <div className="chat-bubble chat-bubble--left">
                    You are in: {room + " room" || "Hall"}
                  </div>
                </div>
              </div>
              <div className="row no-gutters">
                <div className="col-md-3 offset-md-9">
                  {/* <div className="chat-bubble chat-bubble--right">
                    Fell Free to Say!
                  </div> */}
                </div>
              </div>
              <div className="row no-gutters">
                <div className="col-md-3">
                  {/* <div className="chat-bubble chat-bubble--left">
                    Whole World is urs!
                  </div> */}
                </div>
              </div>

              <div>
                {messages.map((txt, index) => (
                        <div key={index} >{txt}</div>
                ))}


{/* {messages.map((txt, index) => (
                  <div key={index} className="row no-gutters">
                    <div className="col-md-3 offset-md-9">
                      <div className="chat-bubble chat-bubble--right">
                        <div>{txt}</div>
                      </div>
                    </div>
                  </div>
                ))} */}
              </div>

              <div className="row">
                <div className="col-12">
                  <div className="chat-box-tray">
                    {/* <i className="material-icons">sentiment_very_satisfied</i> */}
                    <form
                      onSubmit={(e) => {
                        handleSubmitMessage(e);
                      }}
                      className="col-12"
                    >
                      <input
                        type="text"
                        onChange={(e) => setmsg(e.target.value)}
                        placeholder="Type your message here..."
                      />
                      <button type="submit" value="Send" className="col-11">
                        Send
                      </button>
                    </form>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </React.Fragment>
  );
}
