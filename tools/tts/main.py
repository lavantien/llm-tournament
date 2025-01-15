"""
uv init -p 3.12
uv add kokoro-onnx soundfile

wget https://github.com/thewh1teagle/kokoro-onnx/releases/download/model-files/kokoro-v0_19.onnx
wget https://github.com/thewh1teagle/kokoro-onnx/releases/download/model-files/voices.json

uv run main.py
"""

import soundfile as sf
from kokoro_onnx import Kokoro

kokoro = Kokoro("kokoro-v0_19.onnx", "voices.json")
text = """
Mendicants, when it comes to this body made up of the four principal states, an unlearned ordinary person might become disillusioned, dispassionate, and freed. Why is that? This body made up of the four principal states is seen to accumulate and disperse, to be taken up and laid to rest. That’s why, when it comes to this body, an unlearned ordinary person might become disillusioned, dispassionate, and freed.
"""

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
samples, sample_rate = kokoro.create(
    text, voice=VOICES[10], speed=1.0, lang="en-us"
)
sf.write("audio.wav", samples, sample_rate)
print("Created audio.wav")
