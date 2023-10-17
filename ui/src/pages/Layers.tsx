import React from "react";

import NavigationBar from "@/components/navigation/NavigationBar";
import Button from "@/components/buttons/Button";
import Input from "@/components/inputs/Input";
import Dropdown from "@/components/inputs/Dropdown";
import Toggle from "@/components/buttons/Toggle";
import NavigationButton from "@/components/navigation/NavigationButton";
import Card from "@/components/cards/Card";

import SearchIcon from "@/assets/icons/SearchIcon";
import AppsIcon from "@/assets/icons/AppsIcon";
import BarsIcon from "@/assets/icons/BarsIcon";

const Layers: React.FC = () => {
  return (
    <div className="flex bg-primary-100">
      <NavigationBar />
      <div className="relative flex flex-col flex-grow h-screen gap-3 overflow-auto">
        <div className="sticky top-0 flex flex-col p-6 pb-3 gap-6 bg-primary-100">
          <div className="flex justify-between">
            <h1 className="text-[32px] font-extrabold leading-[130%]">
              Layers
            </h1>
            <Button onClick={() => console.log("Refresh layers")}>
              Refresh layers
            </Button>
          </div>
          <Input
            className="w-full"
            placeholder="Search into layers"
            leftIcon={<SearchIcon />}
          />
          <div className="flex flex-row items-center justify-between gap-8">
            <div className="flex flex-row items-center gap-4">
              <span className="text-base font-semibold text-nuances-black">
                267 layers
              </span>
              <span className="border-l h-6 border-primary-600"></span>
              <span className="text-base font-medium text-primary-600">
                Filter by
              </span>
              <div className="flex flex-row items-center gap-2">
                <Dropdown label="State">
                  <></>
                </Dropdown>
                <Dropdown label="Repository">
                  <></>
                </Dropdown>
              </div>
              <div className="flex flex-row items-center gap-[7px]">
                <span className="text-sm font-medium text-nuances-black">
                  Show open PR
                </span>
                <Toggle defaultChecked />
              </div>
            </div>
            <div className="flex flex-row items-center gap-2">
              <NavigationButton icon={<AppsIcon />} />
              <NavigationButton icon={<BarsIcon />} variant="light" />
            </div>
          </div>
        </div>
        <div className="grid grid-cols-[repeat(auto-fit,_minmax(400px,_1fr))] p-6 pt-3 gap-6">
          <Card
            title="fail-terragrunt"
            isRunning
            namespace="burrito-examples"
            state="success"
            repository="burrito"
            branch="failing-terragrunt"
            path="terragrunt/random-pets/test"
            lastResult="error getting last results"
          />
          <Card
            title="fail-terragruntfail-terragruntfail-terragruntfail-terragruntfail-terragruntfail-terragruntfail-terragruntfail-terragrunt"
            namespace="burrito-examples"
            state="success"
            repository="burrito"
            branch="failing-terragrunt"
            path="terragrunt/random-pets/test"
            lastResult="error getting last resultserror getting last resultserror getting last resultserror getting last resultserror getting last results"
          />
          <Card
            title="fail-terragrunt"
            namespace="burrito-examples"
            state="warning"
            repository="burrito"
            branch="failing-terragrunt"
            path="terragrunt/random-pets/test"
            lastResult="error getting last results"
          />
          <Card
            title="fail-terragrunt"
            namespace="burrito-examples"
            state="warning"
            repository="burrito"
            branch="failing-terragrunt"
            path="terragrunt/random-pets/test"
            lastResult="error getting last results"
          />
          <Card
            title="fail-terragrunt"
            namespace="burrito-examples"
            state="error"
            repository="burrito"
            branch="failing-terragrunt"
            path="terragrunt/random-pets/test"
            lastResult="error getting last results"
          />
          <Card
            title="fail-terragrunt"
            namespace="burrito-examples"
            state="success"
            repository="burrito"
            branch="failing-terragrunt"
            path="terragrunt/random-pets/test"
            lastResult="error getting last results"
          />
        </div>
      </div>
    </div>
  );
};

export default Layers;
