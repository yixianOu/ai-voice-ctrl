import React, { useState, useEffect, useRef } from 'react';
import { 
  ProcessVoiceCommandAuto, 
  ProcessTextCommand,
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
  
  // Use ref to track interval for recording status polling
  const pollingIntervalRef = useRef<number | null>(null);
  const errorCountRef = useRef<number>(0);

  // Load available tools on mount
  useEffect(() => {
    loadAvailableTools();
  }, []);

  // Poll recording status continuously while component is mounted.
  // This keeps the UI updated with backend recording state for live reminders.
  useEffect(() => {
    const maxErrors = 5;
    const restartDelayMs = 5000;
    let restartTimeout: number | null = null;

    const checkRecordingStatus = async () => {
      try {
        const recording = await IsRecording();
        setIsRecording(recording);
        errorCountRef.current = 0; // Reset error count on success
      } catch (err) {
        errorCountRef.current++;
        console.warn(`Recording status check failed (${errorCountRef.current}/${maxErrors}):`, err);
        // Stop polling temporarily if too many consecutive errors
        if (errorCountRef.current >= maxErrors && pollingIntervalRef.current !== null) {
          console.error('Too many errors checking recording status, stopping poll temporarily');
          window.clearInterval(pollingIntervalRef.current);
          pollingIntervalRef.current = null;
          setIsRecording(false);
          // Try to restart polling after a backoff
          restartTimeout = window.setTimeout(() => {
            errorCountRef.current = 0;
            if (pollingIntervalRef.current === null) {
              pollingIntervalRef.current = window.setInterval(checkRecordingStatus, 300);
            }
          }, restartDelayMs);
        }
      }
    };

    // start immediate check + interval
    checkRecordingStatus();
    if (pollingIntervalRef.current === null) {
      pollingIntervalRef.current = window.setInterval(checkRecordingStatus, 300);
    }

    return () => {
      if (pollingIntervalRef.current !== null) {
        window.clearInterval(pollingIntervalRef.current);
        pollingIntervalRef.current = null;
      }
      if (restartTimeout !== null) {
        window.clearTimeout(restartTimeout);
        restartTimeout = null;
      }
    };
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
      setResponse(''); // Clear previous response
      setIsProcessing(true);
      
      // This will: record -> transcribe -> LLM -> execute tools
      const result = await ProcessVoiceCommandAuto();
      
      setResponse(result);
      setShowHistory(true);
      await loadConversationHistory();
    } catch (err) {
      console.error('Voice command failed:', err);
      setError(`语音命令失败: ${err}`);
    } finally {
      setIsRecording(false);
      setIsProcessing(false);
    }
  };

  // Text input with LLM processing
  const handleTextCommand = async () => {
    if (!textInput.trim()) return;
    
    const command = textInput; // Save command before clearing
    
    try {
      setError('');
      setResponse(''); // Clear previous response
      setTextInput(''); // Clear input immediately for better UX
      setIsProcessing(true);
      
      const result = await ProcessTextCommand(command);
      setResponse(result);
      setShowHistory(true);
      await loadConversationHistory();
    } catch (err) {
      console.error('Text command failed:', err);
      setError(`文本命令失败: ${err}`);
      setTextInput(command); // Restore input on error
    } finally {
      setIsProcessing(false);
    }
  };

  // Transcribe only (no LLM processing)
  const handleTranscribeOnly = async () => {
    try {
      setError('');
      setTranscript(''); // Clear previous transcript
      setIsProcessing(true);
      
      const text = await RecordAndTranscribe();
      
      // Ensure processing state is cleared before setting text input
      setIsProcessing(false);
      setIsRecording(false);
      setTextInput(text); // Fill into text input for editing
      
      // Focus text input after state updates
      setTimeout(() => {
        const input = document.querySelector('.text-input') as HTMLInputElement;
        if (input) {
          input.focus();
          input.select(); // Select all text for easy editing
        }
      }, 150);
    } catch (err) {
      console.error('Transcription failed:', err);
      setError(`转录失败: ${err}`);
      setIsRecording(false);
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

      {/* Processing Status */}
      {isProcessing && (
        <div className="processing-status">
          <div className="spinner"></div>
          <span>{isRecording ? '🎤 正在录音...' : '⚙️ 处理中...'}</span>
        </div>
      )}

      {/* Recording Status Badge - Always show when recording */}
      {isRecording && (
        <div className="recording-badge">
          <div className="recording-pulse"></div>
          <span>🔴 录音中</span>
        </div>
      )}

      {/* Transcript Display - Only shown when not in text input */}
      {transcript && !textInput && (
        <div className="transcript-box">
          <h3>📝 转录文本</h3>
          <p>{transcript}</p>
          <button 
            className="edit-transcript-button"
            onClick={() => {
              setTextInput(transcript);
              setTranscript('');
              setTimeout(() => {
                const input = document.querySelector('.text-input') as HTMLInputElement;
                if (input) input.focus();
              }, 100);
            }}
          >
            ✏️ 编辑
          </button>
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
