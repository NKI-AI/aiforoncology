# Physician GTV delineation fingerprinting

Quick side project on radiotherapist GTV fingerprinting in an attempt to put together an ESTRO abstract in 5 days. Unfortunately, the signal found in the data was a little too weak to write a proper submission on. Will expand upon this in the near future with more appropriate datasets.

The goal was to train a classifier model that could identify the physician who delineated the GTV on a particular CT scan of early-stage lung cancer patients. The underlying idea was that it is known that inter-observer variation (IOV) exists, but that it is not clear to what extent different observers have their own personal "styles", even though they should all be following guidelines and protocols. Note that such guidelines do not imply that IOV shouldn't exist: there is still some leftover uncertainty and ambiguity, so variations will always exist, but the hope is rather that this variation should be independent of the actual rater.

In the end, two physicians could be singled out a little bit, judging from the model's one-versus-rest AUCs being consistently around 0.6 in a 5-fold crossvalidation. However, the confusion matrix showed that the model still did a fairly bad job overall and it was hard to convince myself that the ~0.6 AUCs were not just due to statistical noise.

Below we see an example of the resulting one-versus-rest AUCs of attempting to classify the responsible physician:

![image](https://github.com/user-attachments/assets/29a4649c-b854-4ffc-a1d9-ee48e7a92ce0)

This looks like there is some kind of signal, but the confusion matrix is less promising:

![image](https://github.com/user-attachments/assets/947670f7-69ce-4180-b1f5-bfd5954c61a5)
