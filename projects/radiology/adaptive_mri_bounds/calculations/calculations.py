def calculate_confusion_matrix(
    sensitivity: float, specificity: float, prevalence: float, population_size: int = 1
) -> dict[str, float]:
    """
    Calculate true positives (TP), false negatives (FN), false positives (FP), and true negatives (TN).

    Parameters
    ----------
    sensitivity : float
        Sensitivity of the test (true positive rate).
    specificity : float
        Specificity of the test (true negative rate).
    prevalence : float
        Proportion of the population with the condition.
    population_size : int
        Total population size. If set to 100, results are percentages.

    Returns
    -------
    dict
        A dictionary containing the counts of true positives ('tp'), false negatives ('fn'),
        false positives ('fp'), and true negatives ('tn').
    """
    true_positives = sensitivity * population_size * prevalence
    false_negatives = (1 - sensitivity) * population_size * prevalence
    false_positives = (1 - specificity) * population_size * (1 - prevalence)
    true_negatives = specificity * population_size * (1 - prevalence)

    return {
        "tp": true_positives,
        "fn": false_negatives,
        "fp": false_positives,
        "tn": true_negatives,
    }
