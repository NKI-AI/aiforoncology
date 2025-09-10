import warnings
from pathlib import Path

from hydra.core.config_search_path import ConfigSearchPath
from hydra.core.plugins import Plugins
from hydra.plugins.search_path_plugin import SearchPathPlugin
from ahcore.exceptions import ConfigurationError
from ahcore.utils.io import get_logger

logger = get_logger(__name__)


class AdditionalSearchPathPlugin(SearchPathPlugin):
    """This plugin allows to overwrite the ahcore configurations without needed to fork the repository."""

    additional_config_path: Path

    @classmethod
    def configure(cls, additional_config_path: Path):
        cls.additional_config_path = additional_config_path
        return cls

    def __init__(self):
        if self.additional_config_path is None:
            raise ConfigurationError("Additional config path not set. Please call configure() first.")

    def manipulate_search_path(self, search_path: ConfigSearchPath) -> None:
        if self.additional_config_path.is_file():
            raise ConfigurationError("Found additional_config file, but expected a folder.")

        elif self.additional_config_path.is_dir():
            if not list(self.additional_config_path.glob("*")):
                warnings.warn(
                    f"Found additional_config folder in {self.additional_config_path}, without any configuration files. "
                    "If you want to overwrite the default ahcore configs, "
                    "please add these to the additional_config folder. "
                    "You can symlink your additional configuration to this folder. "
                    "See the documentation at https://docs.aiforoncology.nl/ahcore/configuration.html "
                    "for more information."
                )
            else:
                # Add additional search path for configs
                logger.info(f"Adding additional search path for configs: file://{self.additional_config_path}")
                search_path.prepend(provider="hydra-ahcore", path=f"file://{self.additional_config_path}")
        else:
            logger.info(
                "No additional_config folder found. Will use standard ahcore configurations."
                "If you want to overwrite or extend the default ahcore configs, you can add these to the "
                "additional_config folder. You could also symlink your additional configuration to this folder."
                "See the documentation at https://docs.aiforoncology.nl/ahcore/configuration.html."
            )


def register_additional_config_search_path(additional_config_path: Path) -> None:
    """
    Register the additional_config folder as a search path for hydra.

    Returns
    -------
    None
    """
    AdditionalSearchPathPlugin.configure(additional_config_path)
    Plugins.instance().register(AdditionalSearchPathPlugin)
