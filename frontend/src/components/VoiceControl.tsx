import React, { useState, useEffect } from 'react';
import { 
  ProcessVoiceCommandAuto, 
  ProcessVoiceCommand,
  RecordAndTranscribe,
  IsRecording,
  GetConversationHistory,
  ResetConversation,
  GetAvailableTools
} from '../../wailsjs/go/main/App';
import { llms } from '../../wailsjs/go/models';
import './VoiceControl.css';

const VoiceControl: React.FC = () => {
  const [isRecording, setIsRecording] = useState(false);
  const [transcript, setTranscript] = useState<string>('');
  const [response, setResponse] = useState<string>('');
  const [error, setError] = useState<string>('');
  const [textInput, setTextInput] = useState<string>('');
  const [conversationHistory, setConversationHistory] = useState<llms.Message[]>([]);
  const [availableTools, setAvailableTools] = useState<string[]>([]);
  const [isProcessing, setIsProcessing] = useState(false);

  // Load available tools on mount
  useEffect(() => {
    loadAvailableTools();
  }, []);

  const loadAvailableTools = async () => {
    try {
      const tools = await GetAvailableTools();
      setAvailableTools(tools);
    } catch (err) {
      console.error('Failed to load tools:', err);
    }
  };

  // Voice input with auto VAD and LLM processing
  const handleVoiceCommand = async () => {
    try {
      setError('');
      setIsProcessing(true);
      setIsRecording(true);
      
      // This will: record -> transcribe -> LLM -> execute tools
      const result = await ProcessVoiceCommandAuto();
      
      setIsRecording(false);
      setResponse(result);
      await loadConversationHistory();
    } catch (err) {
      console.error('Voice command failed:', err);
      setError(`语音命令失败: ${err}`);
      setIsRecording(false);
    } finally {
      setIsProcessing(false);
    }
  };

  // Text input with LLM processing
  const handleTextCommand = async () => {
    if (!textInput.trim()) return;
    
    try {
      setError('');
      setIsProcessing(true);
      
      const result = await ProcessVoiceCommand(textInput);
      setResponse(result);
      setTextInput('');
      await loadConversationHistory();
    } catch (err) {
      console.error('Text command failed:', err);
      setError(`文本命令失败: ${err}`);
    } finally {
      setIsProcessing(false);
    }
  };

  // Transcribe only (no LLM processing)
  const handleTranscribeOnly = async () => {
    try {
      setError('');
      setIsProcessing(true);
      setIsRecording(true);
      
      const text = await RecordAndTranscribe();
      
      setIsRecording(false);
      setTranscript(text);
      setTextInput(text); // Fill into text input for manual processing
    } catch (err) {
      console.error('Transcription failed:', err);
      setError(`转录失败: ${err}`);
      setIsRecording(false);
    } finally {
      setIsProcessing(false);
    }
  };

  const loadConversationHistory = async () => {
    try {
      const history = await GetConversationHistory();
      setConversationHistory(history);
    } catch (err) {
      console.error('Failed to load history:', err);
    }
  };

  const handleResetConversation = async () => {
    try {
      await ResetConversation();
      setConversationHistory([]);
      setResponse('');
      setTranscript('');
      setError('');
    } catch (err) {
      console.error('Failed to reset:', err);
      setError(`重置失败: ${err}`);
    }
  };

  const renderMessage = (msg: llms.Message, index: number) => {
    const roleLabels: Record<string, string> = {
      'system': '系统',
      'user': '用户',
      'assistant': '助手',
      'tool': '工具'
    };

    return (
      <div key={index} className={`message message-${msg.role}`}>
        <div className="message-role">{roleLabels[msg.role] || msg.role}</div>
        <div className="message-content">{msg.content}</div>
        {msg.tool_calls && msg.tool_calls.length > 0 && (
          <div className="tool-calls">
            {msg.tool_calls.map((tc, i) => (
              <div key={i} className="tool-call">
                🔧 {tc.function?.name || 'Unknown tool'}
              </div>
            ))}
          </div>
        )}
      </div>
    );
  };

  return (
    <div className="voice-control">
      <h2>🎤 AI 语音助手</h2>
      
      {/* Main Control Buttons */}
      <div className="control-buttons">
        <button
          className={`primary-button ${isRecording ? 'recording' : ''}`}
          onClick={handleVoiceCommand}
          disabled={isProcessing}
        >
          {isRecording ? '🔴 录音中...' : '🎤 语音命令（自动执行）'}
        </button>
        
        <button
          className="secondary-button"
          onClick={handleTranscribeOnly}
          disabled={isProcessing}
        >
          📝 仅转录（不执行）
        </button>
      </div>

      {/* Text Input */}
      <div className="text-input-container">
        <input
          type="text"
          className="text-input"
          placeholder="输入文本命令..."
          value={textInput}
          onChange={(e) => setTextInput(e.target.value)}
          onKeyPress={(e) => e.key === 'Enter' && handleTextCommand()}
          disabled={isProcessing}
        />
        <button
          className="send-button"
          onClick={handleTextCommand}
          disabled={isProcessing || !textInput.trim()}
        >
          发送
        </button>
      </div>

      {/* Processing Status */}
      {isProcessing && (
        <div className="processing-status">
          <div className="spinner"></div>
          <span>处理中...</span>
        </div>
      )}

      {/* Transcript Display */}
      {transcript && (
        <div className="transcript-box">
          <h3>📝 转录文本：</h3>
          <p>{transcript}</p>
        </div>
      )}

      {/* Response Display */}
      {response && (
        <div className="response-box">
          <h3>🤖 助手响应：</h3>
          <p>{response}</p>
        </div>
      )}

      {/* Error Display */}
      {error && (
        <div className="error-box">
          <p className="error">{error}</p>
        </div>
      )}

      {/* Conversation History */}
      {conversationHistory.length > 0 && (
        <div className="history-container">
          <div className="history-header">
            <h3>💬 对话历史</h3>
            <button className="reset-button" onClick={handleResetConversation}>
              🔄 重置
            </button>
          </div>
          <div className="history-messages">
            {conversationHistory.map((msg, index) => renderMessage(msg, index))}
          </div>
        </div>
      )}

      {/* Available Tools */}
      {availableTools.length > 0 && (
        <div className="tools-container">
          <h3>🔧 可用工具 ({availableTools.length})</h3>
          <div className="tools-list">
            {availableTools.map((tool, index) => (
              <span key={index} className="tool-tag">{tool}</span>
            ))}
          </div>
        </div>
      )}

      {/* Usage Tips */}
      <div className="tips">
        <h3>💡 使用提示：</h3>
        <ul>
          <li><strong>语音命令（自动执行）</strong>：录音 → 转录 → LLM分析 → 自动调用工具</li>
          <li><strong>仅转录</strong>：录音 → 转录到文本框，可手动编辑后发送</li>
          <li><strong>文本输入</strong>：直接输入命令文本，LLM自动调用工具</li>
          <li>示例命令：「创建VSCode工作区 /tmp/test」、「打开hello.md文件」</li>
        </ul>
      </div>
    </div>
  );
};

export default VoiceControl;
