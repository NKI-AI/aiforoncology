#!/bin/bash
#SBATCH --partition a6000
#SBATCH --qos a6000_qos
#SBATCH --gpus 1
#SBATCH --time 1-0:0:0

hostname
echo $SLURM_JOB_ID
echo ${USER}
source /home/${USER}/.bashrc
pwd
source ~/miniconda3/etc/profile.d/conda.sh
conda activate headcal

base_experiment_dir="configs/experiment/train/"

for dataset in $(ls $base_experiment_dir); do
  if [ "$dataset" != "cifar10" ]; then
    continue
  fi
  # join the base experiment with the dataset specific config
  dataset_dir="$base_experiment_dir/$dataset"
  for arch in $(ls $dataset_dir); do
    if [ "$arch" == "pit" ]; then
      continue
    fi
    arch_dir="$dataset_dir/$arch"
    for experiment in $(ls $arch_dir); do
      # trim the .yaml extension
      experiment_name="${experiment%.yaml}"
      experiment_path="train/$dataset/$arch/$experiment_name"
      for seed in {0..4}; do
        srun python src/train.py data.num_workers=14 \
          experiment=$experiment_path seed=$seed +trainer.enable_progress_bar=False \
          model.optimizer.weight_decay=0.5
      done
    done
  done
done
