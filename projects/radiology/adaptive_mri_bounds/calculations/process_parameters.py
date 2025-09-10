from calculations.calculations import calculate_confusion_matrix


def process_parameter_set(**kwargs):
    """
    Process a set of parameters to calculate recall rates and average protocol times.

    Parameters
    ----------
    **kwargs : dict
        Dictionary containing the following parameters:
        sensitivity_full : float
            Sensitivity of the full protocol
        specificity_full : float
            Specificity of the full protocol
        sensitivity_abbr : float
            Sensitivity of the abbreviated protocol
        specificity_abbr : float
            Specificity of the abbreviated protocol
        sensitivity_ai : float
            Sensitivity of the AI model
        specificity_ai : float
            Specificity of the AI model
        prevalence : float
            Disease prevalence in the population
        full_time : float
            Time required for full protocol
        abbr_time : float
            Time required for abbreviated protocol

    Returns
    -------
    dict
        Dictionary containing the following metrics:
        recall_ai_best_case : float
            Best case recall rate with AI
        recall_ai_worst_case : float
            Worst case recall rate with AI
        recall_abbr : float
            Recall rate for abbreviated protocol
        recall_full : float
            Recall rate for full protocol
        avg_time_ai_best_case : float
            Best case average protocol time with AI
        avg_time_ai_worst_case : float
            Worst case average protocol time with AI
        avg_time_abbr : float
            Average protocol time for abbreviated protocol
        avg_time_full : float
            Average protocol time for full protocol
    """
    # Extract parameters
    sensitivity_abbr = kwargs.get("sensitivity_abbr")
    specificity_abbr = kwargs.get("specificity_abbr")
    sensitivity_ai = kwargs.get("sensitivity_ai")
    specificity_ai = kwargs.get("specificity_ai")
    prevalence = kwargs.get("prevalence")
    full_time = kwargs.get("full_time")
    abbr_time = kwargs.get("abbr_time")

    # Calculate true positives (TP), false negatives (FN), false positives (FP), and true negatives (TN)
    abbr_matrix = calculate_confusion_matrix(sensitivity_abbr, specificity_abbr, prevalence)
    ai_matrix = calculate_confusion_matrix(sensitivity_ai, specificity_ai, prevalence)

    # Calculate max overlap (best case scenario)
    max_overlap_tp = min(ai_matrix["tp"], abbr_matrix["tp"])  # Maximum possible overlap
    max_overlap_fp = min(ai_matrix["fp"], abbr_matrix["fp"])
    need_full_tp_best_case = abbr_matrix["tp"] - max_overlap_tp
    need_full_fp_best_case = abbr_matrix["fp"] - max_overlap_fp

    # Calculate min overlap (worst case scenario)
    min_overlap_tp = max(
        0.0, (ai_matrix["tp"] + abbr_matrix["tp"]) - prevalence
    )  # Minimum possible overlap given total positives
    min_overlap_fp = max(
        0.0, (ai_matrix["fp"] + abbr_matrix["fp"]) - (1 - prevalence)
    )  # Minimum possible overlap given total negatives
    need_full_tp_worst_case = abbr_matrix["tp"] - min_overlap_tp
    need_full_fp_worst_case = abbr_matrix["fp"] - min_overlap_fp

    # Calculate recall rate
    recall_rate_best_case = need_full_tp_best_case + need_full_fp_best_case
    recall_rate_worst_case = need_full_tp_worst_case + need_full_fp_worst_case
    recall_rate_abbr = abbr_matrix["tp"] + abbr_matrix["fp"]
    recall_rate_full = 0

    # Calculate average protocol time
    avg_time_best_case = abbr_time * (ai_matrix["tn"] + ai_matrix["fn"]) + full_time * (
        ai_matrix["tp"] + ai_matrix["fp"] + recall_rate_best_case
    )
    avg_time_worst_case = abbr_time * (ai_matrix["tn"] + ai_matrix["fn"]) + full_time * (
        ai_matrix["tp"] + ai_matrix["fp"] + recall_rate_worst_case
    )
    avg_time_abbr = abbr_time + recall_rate_abbr * full_time
    avg_time_full = full_time

    return {
        "recall_ai_best_case": recall_rate_best_case,
        "recall_ai_worst_case": recall_rate_worst_case,
        "recall_abbr": recall_rate_abbr,
        "recall_full": recall_rate_full,
        "avg_time_ai_best_case": avg_time_best_case,
        "avg_time_ai_worst_case": avg_time_worst_case,
        "avg_time_abbr": avg_time_abbr,
        "avg_time_full": avg_time_full,
    }
