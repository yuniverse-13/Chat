// src/components/ChatRoom.js
import React, { useState, useEffect, useRef } from "react";
import './ChatRoom.css';

const ChatRoom = () => {
  const [messages, setMessages] = useState([
    { id: 1, text: 'Привет!', user: 'Enver' },
    { id: 2, text: 'Привет, как дела?', user: 'Yun' },
  ]);
  const [newMessage, setNewMessage] = useState('');
  
  // useRef будет хранить наш WebSocket объект между рендерами.
  const socket = useRef(null);

  useEffect(() => {
    // Устанавливаем соединение при монтировании компонента.
    socket.current = new WebSocket('ws://localhost:8080/ws');

    socket.current.onopen = () => {
      console.log('WebSocket соединение установлено.');
    };

    socket.current.onmessage = (event) => {
      // Пока что сервер отправляет простые строки, а не JSON.
      // Создадим объект сообщения на стороне клиента для отображения.
      const receivedMessage = {
          id: Date.now(), // Генерируем временный ID
          text: event.data,
          user: 'Server Echo' // Указываем, что это эхо от сервера
      };
      setMessages(prevMessages => [...prevMessages, receivedMessage]);
    };

    socket.current.onerror = (error) => {
      console.error('WebSocket ошибка:', error);
    };

    // Функция очистки, которая будет вызвана при размонтировании компонента.
    return () => {
      console.log('WebSocket соединение закрыто.');
      socket.current.close();
    };
  }, []); // Пустой массив зависимостей гарантирует, что эффект выполнится только один раз.

  const handleSendMessage = () => {
    // Проверяем, что сообщение не пустое и сокет готов к работе.
    if (newMessage.trim() !== '' && socket.current && socket.current.readyState === WebSocket.OPEN) {
      // Отправляем текст сообщения на сервер.
      socket.current.send(newMessage);
      setNewMessage('');
    }
  };
  
  return (
    <div className="chat-room">
      <div className="message-list">
        {messages.map(msg => (
          <div key={msg.id} className="message">
            <strong>{msg.user}:</strong> {msg.text}
          </div>
        ))}
      </div>
      <div className="message-input">
        <input
          type="text"
          value={newMessage}
          onChange={(e) => setNewMessage(e.target.value)}
          placeholder="Введите сообщение..."
          // Добавим отправку по нажатию Enter для удобства
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