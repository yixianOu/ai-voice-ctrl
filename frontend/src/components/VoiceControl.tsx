import React, { useState } from 'react';
import { useVoiceRecognition } from '../hooks/useVoiceRecognition';
import { ProcessVoiceCommand } from '../../wailsjs/go/main/App';
import './VoiceControl.css';

const VoiceControl: React.FC = () => {
  const [commandResult, setCommandResult] = useState<string>('');
  const [isProcessing, setIsProcessing] = useState(false);

  // 语音识别结果处理
  const handleVoiceResult = async (result: any) => {
    if (result.isFinal && result.transcript.trim()) {
      console.log('识别结果:', result.transcript);
      console.log('置信度:', result.confidence);

      // 发送到后端处理
      setIsProcessing(true);
      try {
        const response = await ProcessVoiceCommand(result.transcript);
        setCommandResult(response);
      } catch (error) {
        console.error('命令处理失败:', error);
        setCommandResult(`错误: ${error}`);
      } finally {
        setIsProcessing(false);
      }
    }
  };

  const {
    isListening,
    transcript,
    confidence,
    error,
    startListening,
    stopListening,
    isSupported
  } = useVoiceRecognition(handleVoiceResult, 'zh-CN');

  if (!isSupported) {
    return (
      <div className="voice-control">
        <div className="error-message">
          ⚠️ 您的浏览器不支持语音识别功能
        </div>
      </div>
    );
  }

  return (
    <div className="voice-control">
      <h2>🎤 AI 语音助手</h2>
      
      {/* 麦克风按钮 */}
      <div className="mic-container">
        <button
          className={`mic-button ${isListening ? 'listening' : ''}`}
          onClick={isListening ? stopListening : startListening}
          disabled={isProcessing}
        >
          {isListening ? '🔴 停止' : '🎤 开始'}
        </button>
        {isListening && <div className="pulse-ring"></div>}
      </div>

      {/* 状态显示 */}
      <div className="status">
        {isListening && <span className="status-text">正在监听...</span>}
        {isProcessing && <span className="status-text">处理中...</span>}
      </div>

      {/* 实时转录显示 */}
      {transcript && (
        <div className="transcript-box">
          <h3>识别内容：</h3>
          <p className="transcript">{transcript}</p>
          {confidence > 0 && (
            <p className="confidence">置信度: {(confidence * 100).toFixed(1)}%</p>
          )}
        </div>
      )}

      {/* 命令执行结果 */}
      {commandResult && (
        <div className="result-box">
          <h3>执行结果：</h3>
          <p className="result">{commandResult}</p>
        </div>
      )}

      {/* 错误信息 */}
      {error && (
        <div className="error-box">
          <p className="error">{error}</p>
        </div>
      )}

      {/* 使用提示 */}
      <div className="tips">
        <h3>💡 试试说：</h3>
        <ul>
          <li>"打开 Firefox"</li>
          <li>"播放音乐"</li>
          <li>"音量设为 50%"</li>
          <li>"创建文件 test.txt"</li>
        </ul>
      </div>
    </div>
  );
};

export default VoiceControl;
