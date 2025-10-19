import { useState, useEffect, useRef } from 'react';

interface VoiceRecognitionResult {
  transcript: string;
  confidence: number;
  isFinal: boolean;
}

interface UseVoiceRecognitionReturn {
  isListening: boolean;
  transcript: string;
  confidence: number;
  error: string | null;
  startListening: () => void;
  stopListening: () => void;
  isSupported: boolean;
}

/**
 * 语音识别 Hook
 * 使用浏览器的 Web Speech API 进行语音识别
 */
export const useVoiceRecognition = (
  onResult?: (result: VoiceRecognitionResult) => void,
  language: string = 'zh-CN'
): UseVoiceRecognitionReturn => {
  const [isListening, setIsListening] = useState(false);
  const [transcript, setTranscript] = useState('');
  const [confidence, setConfidence] = useState(0);
  const [error, setError] = useState<string | null>(null);
  const recognitionRef = useRef<any>(null);

  // 检查浏览器是否支持语音识别
  const isSupported = 'webkitSpeechRecognition' in window || 'SpeechRecognition' in window;

  useEffect(() => {
    if (!isSupported) {
      setError('您的浏览器不支持语音识别功能');
      return;
    }

    // 创建语音识别实例
    const SpeechRecognition = (window as any).SpeechRecognition || (window as any).webkitSpeechRecognition;
    const recognition = new SpeechRecognition();

    // 配置识别参数
    recognition.continuous = true; // 持续识别
    recognition.interimResults = true; // 返回临时结果
    recognition.lang = language; // 设置语言
    recognition.maxAlternatives = 1; // 最多返回结果数

    // 识别结果处理
    recognition.onresult = (event: any) => {
      let interimTranscript = '';
      let finalTranscript = '';
      let lastConfidence = 0;

      for (let i = event.resultIndex; i < event.results.length; i++) {
        const result = event.results[i];
        const transcriptText = result[0].transcript;

        if (result.isFinal) {
          finalTranscript += transcriptText;
          lastConfidence = result[0].confidence;
          
          // 调用回调
          if (onResult) {
            onResult({
              transcript: transcriptText,
              confidence: lastConfidence,
              isFinal: true
            });
          }
        } else {
          interimTranscript += transcriptText;
        }
      }

      // 更新状态
      const currentTranscript = finalTranscript || interimTranscript;
      setTranscript(currentTranscript);
      setConfidence(lastConfidence);

      // 临时结果也触发回调
      if (!finalTranscript && interimTranscript && onResult) {
        onResult({
          transcript: interimTranscript,
          confidence: 0,
          isFinal: false
        });
      }
    };

    // 错误处理
    recognition.onerror = (event: any) => {
      console.error('语音识别错误:', event.error);
      setError(`识别错误: ${event.error}`);
      setIsListening(false);
    };

    // 识别结束
    recognition.onend = () => {
      setIsListening(false);
    };

    recognitionRef.current = recognition;

    return () => {
      if (recognitionRef.current) {
        recognitionRef.current.stop();
      }
    };
  }, [language, onResult, isSupported]);

  // 开始监听
  const startListening = () => {
    if (!isSupported) {
      setError('您的浏览器不支持语音识别功能');
      return;
    }

    setError(null);
    setTranscript('');
    setConfidence(0);

    try {
      recognitionRef.current?.start();
      setIsListening(true);
    } catch (err: any) {
      console.error('启动识别失败:', err);
      setError(`启动失败: ${err.message}`);
    }
  };

  // 停止监听
  const stopListening = () => {
    try {
      recognitionRef.current?.stop();
      setIsListening(false);
    } catch (err) {
      console.error('停止识别失败:', err);
    }
  };

  return {
    isListening,
    transcript,
    confidence,
    error,
    startListening,
    stopListening,
    isSupported
  };
};
