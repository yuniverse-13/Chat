import React, { useState, useEffect, useRef } from "react";
import './ChatRoom.css';

const ChatRoom = () => {
  const [messages, setMessages] = useState([]);
  const [newMessage, setNewMessage] = useState('');
  
  const [currentUser] = useState(() => "User_" + Math.floor(Math.random() * 1000));

  const socket = useRef(null);

  useEffect(() => {
    socket.current = new WebSocket('ws://localhost:8080/ws');

    socket.current.onopen = () => console.log('WebSocket соединение установлено.');
    socket.current.onerror = (error) => console.error('WebSocket ошибка:', error);
    socket.current.onclose = () => console.log('WebSocket соединение закрыто.');

    socket.current.onmessage = (event) => {
      const receivedMessage = JSON.parse(event.data);
      setMessages(prevMessages => [...prevMessages, receivedMessage]);
    };

    return () => {
      socket.current.close();
    };
  }, []);

  const handleSendMessage = () => {
    if (newMessage.trim() !== '' && socket.current && socket.current.readyState === WebSocket.OPEN) {
      const messageObject = {
        id: Date.now(),
        user: currentUser,
        text: newMessage,
      };

      socket.current.send(JSON.stringify(messageObject));
      setNewMessage('');
    }
  };
  
  return (
    <div className="chat-room">
      <div className="message-list">
        {messages.map(msg => {
          const isCurrentUser = msg.user === currentUser;
          return (
            <div 
              key={msg.id} 
              className={`message ${isCurrentUser ? 'current-user' : ''}`}
            >
              <strong>{msg.user}:</strong> {msg.text}
            </div>
          );
        })}
      </div>
      <div className="message-input">
        <input
          type="text"
          value={newMessage}
          onChange={(e) => setNewMessage(e.target.value)}
          placeholder="Введите сообщение..."
          onKeyPress={(event) => {
            if (event.key === 'Enter') {
              handleSendMessage();
            }
          }}
        />
        <button onClick={handleSendMessage}>Отправить</button>
      </div>
    </div>
  );
};

export default ChatRoom;