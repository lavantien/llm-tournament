"""
Require Cuda 12 and CuDNN 9 installed

uv init -p 3.12
uv add kokoro-onnx soundfile onnxruntime-gpu nvidia-cudnn-cu12

wget https://github.com/thewh1teagle/kokoro-onnx/releases/download/model-files/kokoro-v0_19.onnx
wget https://github.com/thewh1teagle/kokoro-onnx/releases/download/model-files/voices.json

uv run main.py
mpv audio.wav
"""

import soundfile as sf
from kokoro_onnx import Kokoro
from onnxruntime import InferenceSession
from pathlib import Path
import time

start_time = time.time()

ONNX_PROVIDER = "CUDAExecutionProvider"
OUTPUT_FILE = "audio.wav"

TEXT = """
Mendicants, when it comes to this body made up of the four principal states, an unlearned ordinary person might become disillusioned, dispassionate, and freed. Why is that? This body made up of the four principal states is seen to accumulate and disperse, to be taken up and laid to rest. That’s why, when it comes to this body, an unlearned ordinary person might become disillusioned, dispassionate, and freed.
"""
txt = Path('text.txt').read_text(encoding='utf-8')

VOICES = {
    1: 'af_bella',
    2: 'af_nicole',
    3: 'af_sarah',
    4: 'af_sky',
    5: 'am_adam',
    6: 'am_michael',
    7: 'bf_emma',
    8: 'bf_isabella',
    9: 'bm_george',
    10: 'bm_lewis'
}


session = InferenceSession("kokoro-v0_19.onnx", providers=[ONNX_PROVIDER])
kokoro = Kokoro.from_session(session, "voices.json")
samples, sample_rate = kokoro.create(
    txt, voice=VOICES[4], speed=1.0, lang="en-us"
)

sf.write(OUTPUT_FILE, samples, sample_rate)
print("Created audio.wav")

print("--- %s seconds ---" % (time.time() - start_time))
