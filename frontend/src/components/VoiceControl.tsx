import React, { useState, useEffect } from 'react';
import { 
  ProcessVoiceCommandAuto, 
  ProcessVoiceCommand,
  RecordAndTranscribe,
  IsRecording,
  GetConversationHistory,
  ResetConversation,
  GetAvailableTools,
  SetAPIKey
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
  const [showSettings, setShowSettings] = useState(false);
  const [apiKey, setApiKeyInput] = useState<string>('');
  const [showHistory, setShowHistory] = useState(false);

  // Load available tools on mount
  useEffect(() => {
    loadAvailableTools();
  }, []);

  // Poll recording status
  useEffect(() => {
    const checkRecordingStatus = async () => {
      try {
        const recording = await IsRecording();
        setIsRecording(recording);
      } catch (err) {
        // Ignore errors during polling
      }
    };

    // Check recording status every 500ms when processing
    let interval: number | null = null;
    if (isProcessing) {
      interval = window.setInterval(checkRecordingStatus, 500);
    }

    return () => {
      if (interval !== null) window.clearInterval(interval);
    };
  }, [isProcessing]);

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
      
      // This will: record -> transcribe -> LLM -> execute tools
      const result = await ProcessVoiceCommandAuto();
      
      setIsRecording(false);
      setResponse(result);
      setShowHistory(true);
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
      setShowHistory(true);
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
      setShowHistory(false);
    } catch (err) {
      console.error('Failed to reset:', err);
      setError(`重置失败: ${err}`);
    }
  };

  const handleSaveAPIKey = async () => {
    try {
      await SetAPIKey(apiKey);
      setShowSettings(false);
      setApiKeyInput('');
      setError('');
    } catch (err) {
      console.error('Failed to set API key:', err);
      setError(`设置API Key失败: ${err}`);
    }
  };

  const handleQuickCommand = async (command: string) => {
    setTextInput(command);
    // Auto-execute
    try {
      setError('');
      setIsProcessing(true);
      
      const result = await ProcessVoiceCommand(command);
      setResponse(result);
      setShowHistory(true);
      await loadConversationHistory();
    } catch (err) {
      console.error('Quick command failed:', err);
      setError(`命令执行失败: ${err}`);
    } finally {
      setIsProcessing(false);
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
      <div className="header">
        <h2>🎤 AI 语音助手</h2>
        <button 
          className="settings-button"
          onClick={() => setShowSettings(!showSettings)}
          title="设置"
        >
          ⚙️
        </button>
      </div>

      {/* Settings Panel */}
      {showSettings && (
        <div className="settings-panel">
          <h3>⚙️ 设置</h3>
          <div className="setting-item">
            <label>OpenAI API Key:</label>
            <input
              type="password"
              className="api-key-input"
              placeholder="留空则使用环境变量"
              value={apiKey}
              onChange={(e) => setApiKeyInput(e.target.value)}
            />
            <button className="save-button" onClick={handleSaveAPIKey}>
              保存
            </button>
          </div>
        </div>
      )}
      
      {/* Main Control Buttons */}
      <div className="control-buttons">
        <button
          className={`primary-button ${isRecording ? 'recording' : ''}`}
          onClick={handleVoiceCommand}
          disabled={isProcessing}
        >
          {isRecording ? '🔴 录音中...' : '🎤 语音命令'}
        </button>
        
        <button
          className="secondary-button"
          onClick={handleTranscribeOnly}
          disabled={isProcessing}
        >
          📝 仅转录
        </button>
      </div>

      {/* Text Input */}
      <div className="text-input-container">
        <input
          type="text"
          className="text-input"
          placeholder="输入命令或使用语音..."
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

      {/* Quick Commands */}
      <div className="quick-commands">
        <button onClick={() => handleQuickCommand('创建VSCode工作区 /tmp/demo')}>
          📁 创建工作区
        </button>
        <button onClick={() => handleQuickCommand('打开文件 README.md')}>
          📄 打开文件
        </button>
        <button onClick={() => handleQuickCommand('打开浏览器访问 https://github.com')}>
          🌐 打开网页
        </button>
      </div>

      {/* Processing Status */}
      {isProcessing && (
        <div className="processing-status">
          <div className="spinner"></div>
          <span>{isRecording ? '🎤 正在录音...' : '⚙️ 处理中...'}</span>
        </div>
      )}

      {/* Recording Status Badge */}
      {isRecording && !isProcessing && (
        <div className="recording-badge">
          <div className="recording-pulse"></div>
          <span>🔴 录音中</span>
        </div>
      )}

      {/* Transcript Display */}
      {transcript && (
        <div className="transcript-box">
          <h3>📝 转录文本</h3>
          <p>{transcript}</p>
        </div>
      )}

      {/* Response Display */}
      {response && (
        <div className="response-box">
          <h3>🤖 助手响应</h3>
          <p>{response}</p>
        </div>
      )}

      {/* Error Display */}
      {error && (
        <div className="error-box">
          <p className="error">{error}</p>
        </div>
      )}

      {/* Conversation History Toggle */}
      {conversationHistory.length > 0 && (
        <div className="history-toggle">
          <button onClick={() => setShowHistory(!showHistory)}>
            {showHistory ? '▼ 隐藏对话历史' : '▶ 显示对话历史'} ({conversationHistory.length})
          </button>
          <button className="reset-button" onClick={handleResetConversation}>
            🔄 重置对话
          </button>
        </div>
      )}

      {/* Conversation History */}
      {showHistory && conversationHistory.length > 0 && (
        <div className="history-container">
          <div className="history-messages">
            {conversationHistory.map((msg, index) => renderMessage(msg, index))}
          </div>
        </div>
      )}

      {/* Available Tools */}
      <div className="tools-container">
        <h3>🔧 可用工具 ({availableTools.length})</h3>
        <div className="tools-list">
          {availableTools.map((tool, index) => (
            <span key={index} className="tool-tag">{tool}</span>
          ))}
        </div>
      </div>

      {/* Usage Tips */}
      <div className="tips">
        <h3>💡 使用提示</h3>
        <ul>
          <li><strong>语音命令</strong>：点击🎤按钮 → 说话 → 自动转录并执行</li>
          <li><strong>仅转录</strong>：录音后转为文字，可编辑后再发送</li>
          <li><strong>文本输入</strong>：直接输入命令文本执行</li>
          <li><strong>快捷命令</strong>：点击预设按钮快速执行常用命令</li>
        </ul>
      </div>
    </div>
  );
};

export default VoiceControl;
