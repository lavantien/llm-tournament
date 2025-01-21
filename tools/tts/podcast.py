import soundfile as sf
from kokoro_onnx import Kokoro
from onnxruntime import InferenceSession
from pathlib import Path
import time
import numpy as np

start_time = time.time()

ONNX_PROVIDER = "CUDAExecutionProvider"
OUTPUT_FILE = "podcast.wav"

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

# Load podcast script
podcast_script = Path('podcast.txt').read_text(encoding='utf-8').splitlines()

session = InferenceSession("kokoro-v0_19.onnx", providers=[ONNX_PROVIDER])
kokoro = Kokoro.from_session(session, "voices.json")

all_samples = []
sample_rate = None

for line in podcast_script:
    line = line.strip()
    if not line or ':' not in line:
        continue

    speaker, dialogue = line.split(':', 1)
    speaker = speaker.strip()
    dialogue = dialogue.strip()

    if not dialogue:
        continue

    # Select voice based on speaker
    if speaker == 'Lewis':
        voice = VOICES[10]  # bm_lewis
    elif speaker == 'Sky':
        voice = VOICES[4]   # af_sky
    else:
        continue  # Skip unknown speakers

    # Generate audio for this segment
    samples, sr = kokoro.create(dialogue, voice=voice, speed=1.0, lang="en-us")

    if sample_rate is None:
        sample_rate = sr

    all_samples.append(samples)

# Combine all audio segments
if all_samples:
    combined_audio = np.concatenate(all_samples)
    sf.write(OUTPUT_FILE, combined_audio, sample_rate)
    print(f"Successfully created {OUTPUT_FILE}")
else:
    print("No valid dialogue found in podcast.txt")

print("--- %s seconds ---" % (time.time() - start_time))
