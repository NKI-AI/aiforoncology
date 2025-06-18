from pathlib import Path
from typing import Any, Dict, List, Tuple
import dotenv

import hydra
import mlflow
import numpy as np
import torch
from lightning import LightningDataModule, LightningModule, Trainer
from lightning.pytorch.loggers import Logger
from omegaconf import DictConfig
from sklearn.metrics import roc_auc_score
from sklearn.preprocessing import label_binarize
from research.physician_gtv_fingerprint.utils import RankedLogger, extras, instantiate_loggers, task_wrapper

dotenv.load_dotenv(override=True)
log = RankedLogger(__name__, rank_zero_only=True)


@task_wrapper
def evaluate(cfg: DictConfig) -> Tuple[Dict[str, Any], Dict[str, Any]]:
    # Configure MLflow from config
    mlflow.set_tracking_uri(cfg.mlflow.tracking_uri)

    log.info(f"Instantiating model <{cfg.model._target_}>")
    model: LightningModule = hydra.utils.instantiate(cfg.model)

    log.info("Instantiating loggers...")
    logger: List[Logger] = instantiate_loggers(cfg.get("logger"))

    log.info(f"Instantiating trainer <{cfg.trainer._target_}>")
    trainer: Trainer = hydra.utils.instantiate(cfg.trainer, logger=logger)

    experiment = mlflow.get_experiment_by_name(cfg.mlflow.experiment_name)
    experiment_id = experiment.experiment_id

    log_root_dir = Path(cfg.paths.log_dir) / "train/runs"
    log.info("Starting testing!")

    eval_log_dir = Path(trainer.log_dir)
    results_dir = eval_log_dir / "outputs"
    results_dir.mkdir(exist_ok=True)

    fold_aucs = []
    for i in range(5):
        split_label = str(i)
        runs = mlflow.search_runs(
            experiment_ids=experiment_id,
            filter_string=f"params.'data/split_label' = '{split_label}' and tags.mlflow.runName = 'initial'",
        )
        log.info(f"Instantiating datamodule <{cfg.data._target_}>")
        cfg.data.split_label = split_label
        datamodule: LightningDataModule = hydra.utils.instantiate(cfg.data)

        ensemble_probs = []
        targets = None
        for _, run in runs.iterrows():
            ckpt_path = log_root_dir / run["params.log_dir_ckpt"] / "checkpoints" / "epoch_024.ckpt"
            trainer.validate(model=model, datamodule=datamodule, ckpt_path=ckpt_path)
            # apply softmax to the logits
            probs = torch.nn.functional.softmax(model.logits, dim=1).numpy()
            ensemble_probs.append(probs)
            # targets is a numpy array
            if targets is None:
                targets = model.targets
            else:
                assert (targets == model.targets).all()

        # average the probabilities
        ensemble_probs = torch.tensor(ensemble_probs)
        ensemble_probs_mean = ensemble_probs.mean(dim=0)
        ensemble_probs_max = ensemble_probs.max(dim=0).values

        # save probs to a numpy file
        np_probs = ensemble_probs.numpy()
        np_targets = targets
        np.save(results_dir / f"ensemble_probs_f{i}.npy", np_probs)
        np.save(results_dir / f"ensemble_targets_f{i}.npy", np_targets)

        targets_bin = label_binarize(targets, classes=range(model.num_classes))  # type: ignore

        # create mlflow run
        eval_run = mlflow.start_run(experiment_id=experiment_id, run_name="eval")
        # set a parameter for this run
        mlflow.log_param("data/split_label", split_label)
        # save the probs and targets to mlflow
        mlflow.log_artifact(results_dir / f"ensemble_probs_f{i}.npy", run_id=eval_run.info.run_id)
        mlflow.log_artifact(results_dir / f"ensemble_targets_f{i}.npy", run_id=eval_run.info.run_id)

        auc_scores = {}
        auc_scores_max = {}
        for j in range(model.num_classes):
            if targets_bin[:, j].sum() == 0:  # type: ignore
                continue
            auc = roc_auc_score(targets_bin[:, j], ensemble_probs_mean[:, j])  # type: ignore
            auc_max = roc_auc_score(targets_bin[:, j], ensemble_probs_max[:, j])  # type: ignore
            auc_scores[j] = f"{auc:.4f}"
            auc_scores_max[j] = f"{auc_max:.4f}"

            # self.log(f"val/auc_{i}", auc, sync_dist=True, prog_bar=True)  # type: ignore

        fold_aucs.append(auc_scores)

        mlflow.end_run()

        print("-" * 80)
        print(f"fold {i} aucs: {auc_scores}")
        print(f"fold {i} max aucs: {auc_scores_max}")
        print("-" * 80)

    print(f"fold aucs: {fold_aucs}")
    # average the aucs
    avg_aucs = {}
    for i in range(model.num_classes):
        aucs = [float(fold[i]) for fold in fold_aucs if i in fold]
        avg_aucs[i] = sum(aucs) / len(aucs)

    return {}, {}


@hydra.main(version_base="1.3", config_path="configs", config_name="eval.yaml")
def main(cfg: DictConfig) -> None:
    """Main entry point for evaluation.

    :param cfg: DictConfig configuration composed by Hydra.
    """
    # apply extra utilities
    # (e.g. ask for tags if none are provided in cfg, print cfg tree, etc.)
    extras(cfg)

    evaluate(cfg)


if __name__ == "__main__":
    main()
