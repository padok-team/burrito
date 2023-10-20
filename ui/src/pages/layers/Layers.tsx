import React, { useState, useContext, useEffect } from "react";

import { ThemeContext } from "@/contexts/ThemeContext";

import NavigationBar from "@/components/navigation/NavigationBar";
import Button from "@/components/buttons/Button";
import Input from "@/components/inputs/Input";
import Dropdown from "@/components/inputs/Dropdown";
import Toggle from "@/components/buttons/Toggle";
import NavigationButton from "@/components/navigation/NavigationButton";
import Card from "@/components/cards/Card";
import Table from "@/components/tables/Table";

import StateDropdown from "@/pages/layers/components/StateDropdown";
import RepositoryDropdown from "@/pages/layers/components/RepositoryDropdown";

import { Layer, LayerState } from "@/types/types";

import SearchIcon from "@/assets/icons/SearchIcon";
import AppsIcon from "@/assets/icons/AppsIcon";
import BarsIcon from "@/assets/icons/BarsIcon";

const Layers: React.FC = () => {
  const testData: Layer[] = [
    {
      namespace: "burrito-examples",
      name: "fail-terragrunt",
      state: "success",
      repository: "burrito-1",
      branch: "failling-terraform",
      path: "terragrunt/random-pets/test",
      lastResult: "error getting last results",
      isRunning: false,
      isPR: true,
    },
    {
      namespace: "burrito-examples",
      name: "fail-terragrunt",
      state: "warning",
      repository: "burrito-2",
      branch: "failling-terraform",
      path: "terragrunt/random-pets/test",
      lastResult: "error getting last results",
      isRunning: true,
      isPR: false,
    },
    {
      namespace: "burrito-examples",
      name: "fail-terragrunt",
      state: "error",
      repository: "burrito-3",
      branch: "failling-terraform",
      path: "terragrunt/random-pets/test",
      lastResult: "error getting last results",
      isRunning: false,
      isPR: true,
    },
  ];

  const { theme } = useContext(ThemeContext);
  const [view, setView] = useState<"grid" | "table">("grid");
  const [stateFilter, setStateFilter] = useState<LayerState[]>([]);
  const [repositoryFilter, setRepositoryFilter] = useState<string[]>([]);
  const [displayOnlyPRFilter, setDisplayOnlyPRFilter] =
    useState<boolean>(false);
  const [data] = useState<Layer[]>(testData); // TODO: replace with data from API
  const [filteredData, setFilteredData] = useState<Layer[]>([]);

  useEffect(() => {
    setFilteredData(
      data
        .filter((layer) =>
          stateFilter.length === 0 ? layer : stateFilter.includes(layer.state)
        )
        .filter((layer) =>
          repositoryFilter.length === 0
            ? layer
            : repositoryFilter.includes(layer.repository)
        )
        .filter((layer) => !displayOnlyPRFilter || layer.isPR)
    );
  }, [data, stateFilter, repositoryFilter, displayOnlyPRFilter]);

  return (
    <div
      className={`
        flex
        ${theme === "light" ? "bg-primary-100" : "bg-nuances-black"}
      `}
    >
      <NavigationBar variant={theme} />
      <div className="relative flex flex-col flex-grow h-screen gap-3 overflow-auto">
        <div
          className={`
            sticky
            top-0
            z-10
            flex
            flex-col
            p-6
            pb-3
            gap-6
            ${theme === "light" ? "bg-primary-100" : "bg-nuances-black"}
          `}
        >
          <div className="flex justify-between">
            <h1
              className={`
                text-[32px]
                font-extrabold
                leading-[130%]
                ${theme === "light" ? "text-nuances-black" : "text-nuances-50"}
              `}
            >
              Layers
            </h1>
            <Button
              variant={theme === "light" ? "primary" : "secondary"}
              onClick={() => console.log("Refresh layers")}
            >
              Refresh layers
            </Button>
          </div>
          <Input
            variant={theme}
            className="w-full"
            placeholder="Search into layers"
            leftIcon={<SearchIcon />}
          />
          <div className="flex flex-row items-center justify-between gap-8">
            <div className="flex flex-row items-center gap-4">
              <span
                className={`
                  text-base
                  font-semibold
                  ${
                    theme === "light" ? "text-nuances-black" : "text-nuances-50"
                  }
              `}
              >
                267 layers
              </span>
              <span
                className={`
                  border-l
                  h-6
                  ${
                    theme === "light"
                      ? "border-primary-600"
                      : "border-nuances-200"
                  }
                `}
              ></span>
              <span
                className={`
                  text-base
                  font-medium
                  ${theme === "light" ? "text-primary-600" : "text-nuances-200"}
              `}
              >
                Filter by
              </span>
              <div className="flex flex-row items-center gap-2">
                <Dropdown
                  variant={theme}
                  label="State"
                  filled={stateFilter.length !== 0}
                >
                  <StateDropdown
                    variant={theme}
                    filter={stateFilter}
                    onChange={setStateFilter}
                  />
                </Dropdown>
                <Dropdown
                  variant={theme}
                  label="Repository"
                  filled={repositoryFilter.length !== 0}
                >
                  <RepositoryDropdown variant={theme} />
                </Dropdown>
              </div>
              <div className="flex flex-row items-center gap-[7px]">
                <span
                  className={`
                    text-sm
                    font-medium
                    ${
                      theme === "light"
                        ? "text-nuances-black"
                        : "text-nuances-50"
                    }
                  `}
                >
                  Show only open PR
                </span>
                <Toggle
                  checked={displayOnlyPRFilter}
                  onChange={() => setDisplayOnlyPRFilter(!displayOnlyPRFilter)}
                />
              </div>
            </div>
            <div className="flex flex-row items-center gap-2">
              <NavigationButton
                icon={<AppsIcon />}
                variant={theme}
                selected={view === "grid"}
                onClick={() => setView("grid")}
              />
              <NavigationButton
                icon={<BarsIcon />}
                variant={theme}
                selected={view === "table"}
                onClick={() => setView("table")}
              />
            </div>
          </div>
        </div>
        {view === "grid" && (
          <div className="grid grid-cols-[repeat(auto-fit,_minmax(400px,_1fr))] p-6 pt-3 gap-6">
            {filteredData.map((layer, index) => (
              <Card key={index} variant={theme} layer={layer} />
            ))}
          </div>
        )}
        {view === "table" && <Table variant={theme} data={filteredData} />}
      </div>
    </div>
  );
};

export default Layers;
