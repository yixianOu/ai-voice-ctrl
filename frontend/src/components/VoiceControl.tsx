import React, { useState } from 'react';
import { StartRecording, StopRecording, ProcessVoiceCommand } from '../../wailsjs/go/main/App';
import './VoiceControl.css';

const VoiceControl: React.FC = () => {
  const [isRecording, setIsRecording] = useState(false);
  const [audioFile, setAudioFile] = useState<string>('');
  const [commandResult, setCommandResult] = useState<string>('');
  const [error, setError] = useState<string>('');

  // 开始录音
  const handleStartRecording = async () => {
    try {
      setError('');
      const filePath = await StartRecording();
      setAudioFile(filePath);
      setIsRecording(true);
      console.log('录音已开始，文件:', filePath);
    } catch (err) {
      console.error('启动录音失败:', err);
      setError(`启动录音失败: ${err}`);
    }
  };

  // 停止录音
  const handleStopRecording = async () => {
    try {
      const result = await StopRecording();
      setIsRecording(false);
      console.log('录音已停止:', result);
      
      // 显示录音文件路径
      setCommandResult(`录音已保存: ${audioFile}`);
    } catch (err) {
      console.error('停止录音失败:', err);
      setError(`停止录音失败: ${err}`);
    }
  };

  // 处理命令（用于文本输入测试）
  const handleTestCommand = async (command: string) => {
    try {
      const result = await ProcessVoiceCommand(command);
      setCommandResult(result);
    } catch (err) {
      console.error('命令执行失败:', err);
      setError(`命令执行失败: ${err}`);
    }
  };

  return (
    <div className="voice-control">
      <h2>🎤 AI 语音助手</h2>
      
      {/* 录音按钮 */}
      <div className="mic-container">
        <button
          className={`mic-button ${isRecording ? 'listening' : ''}`}
          onClick={isRecording ? handleStopRecording : handleStartRecording}
        >
          {isRecording ? '🔴 停止录音' : '🎤 开始录音'}
        </button>
        {isRecording && <div className="pulse-ring"></div>}
      </div>

      {/* 状态显示 */}
      <div className="status">
        {isRecording && <span className="status-text">正在录音中...</span>}
        {audioFile && !isRecording && <span className="status-text">录音文件: {audioFile}</span>}
      </div>

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

      {/* 测试命令输入 */}
      <div className="tips">
        <h3>💡 测试命令：</h3>
        <div style={{ display: 'flex', gap: '10px', marginTop: '10px', flexWrap: 'wrap' }}>
          <button onClick={() => handleTestCommand('打开火狐')}>打开火狐</button>
          <button onClick={() => handleTestCommand('音量设为50%')}>音量50%</button>
          <button onClick={() => handleTestCommand('播放音乐')}>播放音乐</button>
          <button onClick={() => handleTestCommand('下一首')}>下一首</button>
        </div>
        <p style={{ marginTop: '20px', fontSize: '14px', color: '#666' }}>
          提示：点击"开始录音"按钮后，会录制5秒音频。录音完成后需要手动实现音频转文字功能。
        </p>
      </div>
    </div>
  );
};

export default VoiceControl;
