# Background Batch Eraser

`uv init`

`uv venv`

`.venv/Scripts/activate`

`uv pip install torch torchvision --default-index https://download.pytorch.org/whl/cu126`

`uv add torch torchvision`

`uv run python -c 'import torch; print(torch.__version__)'`

`uv add numpy einops Pillow timm opencv-python`

`uv run main.py`

Requires `ffmpeg` for video segmentation

<https://huggingface.co/PramaLLC/BEN2>
