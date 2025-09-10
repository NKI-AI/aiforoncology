from pathlib import Path
from typing import Dict, List
import dotenv

import hydra
import mlflow
import numpy as np
from matplotlib import pyplot as plt
from omegaconf import DictConfig
from sklearn.metrics import (
    ConfusionMatrixDisplay,
    accuracy_score,
    confusion_matrix,
    roc_auc_score,
    top_k_accuracy_score,
)
from sklearn.preprocessing import label_binarize
from research.physician_gtv_fingerprint.utils import RankedLogger, extras

dotenv.load_dotenv(override=True)
log = RankedLogger(__name__, rank_zero_only=True)


def plot_class_aucs(aucs_ensemble: Dict[int, List[float]]) -> None:
    """Plot AUC scores for each class.

    :param aucs_ensemble: Dictionary mapping class indices to lists of AUC scores
    """
    fig, ax = plt.subplots()

    # make a scatter plot of the aucs
    for i, aucs in aucs_ensemble.items():
        ax.scatter([i] * len(aucs), aucs, label=f"Class {i}", alpha=0.5)
        ax.scatter(i, np.mean(aucs), color="black", marker="x")

    # draw a horizontal line at 0.5
    ax.axhline(0.5, color="black", linestyle="--", alpha=0.5)

    ax.set_xlabel("Class")
    ax.set_ylabel("AUC")
    ax.set_ylim(0, 1.0)
    ax.set_xticks(range(10))
    plt.show()


def plot_accuracies(accs_ensemble: List[float], topk_accs_ensemble: List[float]) -> None:
    """Plot accuracy scores.

    :param accs_ensemble: List of accuracy scores
    :param topk_accs_ensemble: List of top-k accuracy scores
    """
    fig, ax = plt.subplots()

    ax.scatter([0] * len(accs_ensemble), accs_ensemble, alpha=0.5)
    ax.scatter([1] * len(topk_accs_ensemble), topk_accs_ensemble, alpha=0.5)

    ax.set_xlabel("Metric")
    ax.set_ylabel("Accuracy")
    ax.set_ylim(0, 1.0)
    ax.set_xticks([0, 1])
    ax.set_xticklabels(["Top-1", "Top-3"])
    plt.show()


def analyze(cfg: DictConfig) -> None:
    """Analyze the results of the ensemble model.

    :param cfg: Configuration composed by Hydra
    """
    # Configure MLflow
    mlflow.set_tracking_uri(cfg.mlflow.tracking_uri)
    experiment = mlflow.get_experiment_by_name(cfg.mlflow.experiment_name)
    if experiment is None:
        raise ValueError(f"Experiment {cfg.mlflow.experiment_name} not found")
    experiment_id = experiment.experiment_id

    # Get all runs in this experiment with name 'eval'
    eval_runs = mlflow.search_runs(experiment_ids=experiment_id, filter_string="tags.mlflow.runName = 'eval'")

    aucs_single = {}
    aucs_ensemble = {i: [] for i in range(10)}
    accs_ensemble = []
    topk_accs_ensemble = []

    for _, run in eval_runs.iterrows():
        split_label = run["params.data/split_label"]
        artifact_uri = Path(run.artifact_uri)

        ensemble_probs = np.load(artifact_uri / f"ensemble_probs_f{split_label}.npy")
        fold_targets = np.load(artifact_uri / f"ensemble_targets_f{split_label}.npy")

        targets_bin = label_binarize(fold_targets, classes=np.unique(fold_targets))

        for j in range(5):
            for k in range(ensemble_probs.shape[-1]):
                if targets_bin[:, k].sum() == 0:
                    continue
                auc = roc_auc_score(targets_bin[:, k], ensemble_probs[j, :, k])

        ensemble_probs_mean = ensemble_probs.mean(axis=0)
        for k in range(ensemble_probs_mean.shape[-1]):
            if targets_bin[:, k].sum() == 0:
                continue
            auc_mean = roc_auc_score(targets_bin[:, k], ensemble_probs_mean[:, k])
            aucs_ensemble[k].append(auc_mean)

        acc = accuracy_score(fold_targets, ensemble_probs_mean.argmax(axis=1))
        accs_ensemble.append(acc)

        topk_acc = top_k_accuracy_score(fold_targets, ensemble_probs_mean, k=3)
        topk_accs_ensemble.append(topk_acc)

        # Construct and plot the confusion matrix
        cm = confusion_matrix(fold_targets, ensemble_probs_mean.argmax(axis=1))
        disp = ConfusionMatrixDisplay(confusion_matrix=cm)
        disp.plot(cmap="Blues")

    # Plot the results
    plot_class_aucs(aucs_ensemble)
    plot_accuracies(accs_ensemble, topk_accs_ensemble)


@hydra.main(version_base="1.3", config_path="configs", config_name="analyze.yaml")
def main(cfg: DictConfig) -> None:
    """Main entry point for analysis.

    :param cfg: DictConfig configuration composed by Hydra.
    """
    # Apply extra utilities
    extras(cfg)

    analyze(cfg)


if __name__ == "__main__":
    main()
