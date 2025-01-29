import BEN2
import torch


device = torch.device('cuda' if torch.cuda.is_available() else 'cpu')

video_path = "video.mp4"  # input video

model = BEN2.BEN_Base().to(device).eval()  # init pipeline

model.loadcheckpoints("./BEN2_Base.pth")


model.segment_video(
    video_path=video_path,
    # Outputs will be saved as foreground.webm or foreground.mp4. The default value is "./"
    output_path="./",
    # If this is set to 0 CV2 will detect the fps in the original video. The default value is 0.
    fps=0,
    # refine foreground is an extract postprocessing step that increases inference time but can improve matting edges. The default value is False.
    refine_foreground=True,
    # We recommended that batch size not exceed 3 for consumer GPUs as there are minimal inference gains. The default value is 1.
    batch=1,
    # Informs you what frame is being processed. The default value is True.
    print_frames_processed=True,
    # This will output an alpha layer video but this defaults to mp4 when webm is false. The default value is False.
    webm=True,
    # If you do not use webm this will be the RGB value of the resulting background only when webm is False. The default value is a green background (0,255,0).
    rgb_value=(0, 255, 0)
)
