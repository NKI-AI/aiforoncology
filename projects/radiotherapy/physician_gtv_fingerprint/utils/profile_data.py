import cProfile
import pstats

import torch
from research.physician_gtv_fingerprint.data.ct_datamodule import CTDataModule
from tqdm import tqdm

dm = CTDataModule(batch_size=8)

dm.setup()

dl_train = dm.train_dataloader()


cp = cProfile.Profile()
cp.enable()

print("starting...")
nSlices = []
for i, batch in tqdm(enumerate(dl_train)):
    breakpoint()

    if i == 10:
        break

cp.disable()

pstats.Stats(cp).sort_stats("cumulative").print_stats(10)

nSlices = torch.cat(nSlices)

print(nSlices)
