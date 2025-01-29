import BEN2
from PIL import Image
import torch

print(torch.cuda.is_available())
device = torch.device('cuda' if torch.cuda.is_available() else 'cpu')


model = BEN2.BEN_Base().to(device).eval()  # init pipeline

model.loadcheckpoints("./BEN2_Base.pth")

file1 = "./image1.png"  # input image1
file2 = "./image2.png"  # input image2
file3 = "./image3.png"  # input image2
image1 = Image.open(file1)
image2 = Image.open(file2)
image3 = Image.open(file3)


# We recommend that the batch size not exceed 3 for consumer GPUs as there are minimal inference gains due to our custom batch processing for the MVANet decoding steps.
foregrounds = model.inference([image1, image2, image3], refine_foreground=True)
foregrounds[0].save("./foreground1.png")
foregrounds[1].save("./foreground2.png")
foregrounds[2].save("./foreground3.png")
