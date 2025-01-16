# Local TTS

Require Cuda 12 and CuDNN 9 installed

```bash
uv init -p 3.12
```

```bash
uv add kokoro-onnx soundfile onnxruntime-gpu nvidia-cudnn-cu12
```

```bash
wget https://github.com/thewh1teagle/kokoro-onnx/releases/download/model-files/kokoro-v0_19.onnx
```

```bash
wget https://github.com/thewh1teagle/kokoro-onnx/releases/download/model-files/voices.json
```

```bash
uv run main.py
```

```bash
mpv audio.wav
```
