import React, { useState, useContext, useCallback, useMemo } from "react";
import { useSearchParams } from "react-router-dom";
import { useInfiniteQuery, InfiniteQueryObserverSuccessResult } from "@tanstack/react-query";

import InfiniteScroll from 'react-infinite-scroll-component';

import { fetchLayers } from "@/clients/layers/client";
import { reactQueryKeys } from "@/clients/reactQueryConfig";

import { ThemeContext } from "@/contexts/ThemeContext";

import Button from "@/components/core/Button";
import Input from "@/components/core/Input";
import StatesDropdown from "@/components/dropdowns/StatesDropdown";
import RepositoriesDropdown from "@/components/dropdowns/RepositoriesDropdown";
import Toggle from "@/components/core/Toggle";
import NavigationButton from "@/components/navigation/NavigationButton";
import Card from "@/components/cards/Card";
import Table from "@/components/tables/Table";

import SearchIcon from "@/assets/icons/SearchIcon";
import AppsIcon from "@/assets/icons/AppsIcon";
import BarsIcon from "@/assets/icons/BarsIcon";
import CardLoader from "@/components/loaders/CardLoader";

import { LayerState } from "@/clients/layers/types";

import { Layer } from "@/clients/layers/types";
import useFixMissingScroll from "@/components/misc/FixMissingScroll";

const Layers: React.FC = () => {
  const { theme } = useContext(ThemeContext);
  const [view, setView] = useState<"grid" | "table">("grid");
  const [searchParams, setSearchParams] = useSearchParams();
  const [next] = useState<string | undefined>(undefined);
  const search = useMemo<string>(
    () => {
      return searchParams.get("search") || ""
    },
    [searchParams]
  );

  const setSearch = useCallback(
    (search: string) => {
      searchParams.set("search", search);
      setSearchParams(searchParams);
    },
    [searchParams, setSearchParams]
  );

  const stateFilter = useMemo<LayerState[]>(() => {
    const param = searchParams.get("states");
    return (param ? param.split(",") : []) as LayerState[];
  }, [searchParams]);

  const setStateFilter = useCallback(
    (stateFilter: LayerState[]) => {
      searchParams.set("states", stateFilter.join(","));
      setSearchParams(searchParams);
    },
    [searchParams, setSearchParams]
  );

  const repositoryFilter = useMemo<string[]>(() => {
    const param = searchParams.get("repositories");
    return param ? param.split(",") : [];
  }, [searchParams]);

  const setRepositoryFilter = useCallback(
    (repositoryFilter: string[]) => {
      searchParams.set("repositories", repositoryFilter.join(","));
      setSearchParams(searchParams);
    },
    [searchParams, setSearchParams]
  );

  const hidePRFilter = useMemo<boolean>(
    () => searchParams.get("hidepr") !== "false",
    [searchParams]
  );

  const setHidePRFilter = useCallback(
    (hidePRFilter: boolean) => {
      searchParams.set("hidepr", hidePRFilter.toString());
      setSearchParams(searchParams);
    },
    [searchParams, setSearchParams]
  );
  
const applyResultFilters = (result: InfiniteQueryObserverSuccessResult<{
    results: Layer[];
    pageParams: (string | undefined)[];
}, Error>) => {
  return result.data.results.filter((layer) =>
    layer.name.toLowerCase().includes(search.toLowerCase())
  ).filter((layer) =>
    stateFilter.length === 0 || stateFilter.includes(layer.state)
  ).filter((layer) =>
    repositoryFilter.length === 0 || repositoryFilter.includes(layer.repository)
  ).filter((layer) => !hidePRFilter || !layer.isPR)
}


  const layersQuery = useInfiniteQuery({
    queryKey: reactQueryKeys.layers(9, next),
    queryFn: async ({pageParam}) => await fetchLayers(9, pageParam),
    initialPageParam: next,
    getNextPageParam: (lastPage) => lastPage.next,
    select: data => ({
      results: data.pages.flatMap(layers =>  layers.results),
      pageParams: data.pageParams
    })
    
    // (data) => ({
    //   pages: data.pages
    //     .filter((page) =>
    //       page.results.filter(layer => layer.name.toLowerCase().includes(search.toLowerCase()))
    //     )
    //     .filter((page) =>
    //       page.results.filter((layer) =>
    //         stateFilter.length === 0 || stateFilter.includes(layer.state))
    //     )
    //     .filter(
    //       (page) =>
    //         page.results.filter((layer) =>
    //           repositoryFilter.length === 0 ||
    //           repositoryFilter.includes(layer.repository)
    //         )
    //     )
    //     .filter((page) => 
    //       page.results.filter(layer => !hidePRFilter || !layer.isPR)),
    // }),
});

  useFixMissingScroll({
    hasMoreItems: layersQuery.hasNextPage,
    fetchMoreItems: () => layersQuery.fetchNextPage(),
    mainElement: document.getElementById("scrollableDiv") as HTMLElement
  });

  return (
    <div className="flex flex-col flex-1 h-screen min-w-0">
      <div
        className={`
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
            isLoading={layersQuery.isRefetching}
            onClick={() => layersQuery.refetch()}
          >
            Refresh layers
          </Button>
        </div>
        <Input
          variant={theme}
          className="w-full"
          placeholder="Search into layers"
          leftIcon={<SearchIcon />}
          value={search}
          onChange={(e) => setSearch(e.target.value)}
        />
        <div className="flex flex-row items-center justify-between gap-8">
          <div className="flex flex-row items-center gap-4">
            <span
              className={`
                text-base
                font-semibold
                ${theme === "light" ? "text-nuances-black" : "text-nuances-50"}
              `}
            >
              {`
                ${
                  layersQuery.isSuccess ? layersQuery.data.results.length : 0
                } layers loaded
              `}
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
              <StatesDropdown
                variant={theme}
                selectedStates={stateFilter}
                setSelectedStates={setStateFilter}
              />
              <RepositoriesDropdown
                variant={theme}
                selectedRepositories={repositoryFilter}
                setSelectedRepositories={setRepositoryFilter}
              />
            </div>
            <Toggle
              className={`
                text-sm
                font-medium
                ${theme === "light" ? "text-nuances-black" : "text-nuances-50"}
              `}
              checked={hidePRFilter}
              onChange={() => setHidePRFilter(!hidePRFilter)}
              label="Hide Pull Requests"
            />
          </div>
          <div className="flex flex-row items-center gap-8">
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
      </div>
      <div
        id="scrollableDiv"
        className={`
          relative
          ${layersQuery.isSuccess ? "overflow-auto" : "overflow-hidden"}
        `}
      >
        {view === "grid" ? (
          <div className="grid grid-cols-[repeat(auto-fit,_minmax(400px,_1fr))] p-6 gap-6">
            {layersQuery.isLoading ? (
              Array.from({ length: 100 }).map((_, index) => (
                <CardLoader key={index} variant={theme} />
              ))
            ) : layersQuery.isError ? (
              <span
                className={`
                  text-lg
                  font-semibold
                  ${
                    theme === "light" ? "text-nuances-black" : "text-nuances-50"
                  }
                `}
              >
                An error has occurred.
              </span>
            ) : layersQuery.isSuccess ? 
            (
              layersQuery.data.results.length > 0 ?
                <InfiniteScroll
                dataLength={layersQuery.data.results.length}
                next={layersQuery.fetchNextPage}
                hasMore={layersQuery.hasNextPage}
                loader={<CardLoader variant={theme} />}
                endMessage={<p>No more data to load.</p>}
                scrollableTarget="scrollableDiv"
                className="grid grid-cols-[repeat(auto-fit,_minmax(400px,_1fr))] p-6 gap-6"
              >
                  {applyResultFilters(layersQuery).map((layer, layerIndex) => <Card key={layerIndex+1} variant={theme} layer={layer} /> )}
              </InfiniteScroll>
                : (
                <span
                  className={`
                    text-lg
                    font-semibold
                    ${
                      theme === "light"
                        ? "text-nuances-black"
                        : "text-nuances-50"
                    }
                  `}
                >
                  No layers found
                </span>
              )
            ) : (
              <></>
            )}
          </div>
        ) : view === "table" ? (
          <div>
            {layersQuery.isLoading ? (
              <Table variant={theme} isLoading data={[]} />
            ) : layersQuery.isError ? (
              <span
                className={`
                  text-lg
                  font-semibold
                  ${
                    theme === "light" ? "text-nuances-black" : "text-nuances-50"
                  }
                `}
              >
                An error has occurred.
              </span>
            ) : layersQuery.isSuccess ? (
              layersQuery.data.results.length > 0 ? (
                <InfiniteScroll
                dataLength={layersQuery.data.results.length}
                next={layersQuery.fetchNextPage}
                hasMore={layersQuery.hasNextPage}
                loader={<Table variant={theme} isLoading data={[]} />}
                endMessage={<p>No more data to load.</p>}
                scrollableTarget="scrollableDiv"
              >
                  <Table variant={theme} data={applyResultFilters(layersQuery)} />
              </InfiniteScroll>
                // <Table variant={theme} data={layersQuery.data.results.reduce((acc, page) => acc.concat(page.results), [] as Layer[])} />
              ) : (
                <div className="p-6">
                  <span
                    className={`
                    text-lg
                    font-semibold
                    ${
                      theme === "light"
                        ? "text-nuances-black"
                        : "text-nuances-50"
                    }
                  `}
                  >
                    No layers found
                  </span>
                </div>
              )
            ) : (
              <></>
            )}
          </div>
        ) : (
          <></>
        )}
      </div>
    </div>
  );
};

export default Layers;
