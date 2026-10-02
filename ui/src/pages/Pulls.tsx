import React, { useContext, useMemo, useState } from 'react';
import { useSearchParams } from 'react-router';
import { useQuery } from '@tanstack/react-query';

import { fetchPullRequests } from '@/clients/pulls/client';
import { reactQueryKeys } from '@/clients/reactQueryConfig';
import { Layer } from '@/clients/layers/types';
import { PullRequest } from '@/clients/pulls/types';

import { ThemeContext } from '@/contexts/ThemeContext';

import Button from '@/components/core/Button';
import Input from '@/components/core/Input';
import RepositoriesDropdown from '@/components/dropdowns/RepositoriesDropdown';
import LogsTerminal from '@/components/tools/LogsTerminal';
import Tag from '@/components/widgets/Tag';
import PullRequestStateTag from '@/components/widgets/PullRequestStateTag';

import SearchIcon from '@/assets/icons/SearchIcon';
import CodeBranchIcon from '@/assets/icons/CodeBranchIcon';

const prKey = (pr: PullRequest) => `${pr.namespace}/${pr.name}`;
const layerKey = (layer: Layer) => `${layer.namespace}/${layer.name}`;

const Pulls: React.FC = () => {
  const { theme } = useContext(ThemeContext);
  const [searchParams, setSearchParams] = useSearchParams();
  const [selectedLayer, setSelectedLayer] = useState<string | null>(null);

  const search = searchParams.get('search') || '';
  const selectedPR = searchParams.get('pr');
  const repositoryFilter = useMemo<string[]>(() => {
    const param = searchParams.get('repositories');
    return param ? param.split(',') : [];
  }, [searchParams]);

  const updateParam = (key: string, value: string) => {
    const next = new URLSearchParams(searchParams);
    if (value) {
      next.set(key, value);
    } else {
      next.delete(key);
    }
    setSearchParams(next);
  };

  const pullsQuery = useQuery({
    queryKey: reactQueryKeys.pulls,
    queryFn: fetchPullRequests
  });

  const filteredPulls = useMemo(() => {
    const term = search.toLowerCase();
    return (pullsQuery.data?.results ?? [])
      .filter(
        (pr) =>
          pr.id.toLowerCase().includes(term) ||
          pr.name.toLowerCase().includes(term) ||
          pr.branch.toLowerCase().includes(term)
      )
      .filter(
        (pr) =>
          repositoryFilter.length === 0 ||
          repositoryFilter.includes(pr.repository)
      );
  }, [pullsQuery.data?.results, search, repositoryFilter]);

  const currentPR = filteredPulls.find((pr) => prKey(pr) === selectedPR);
  const currentLayer = currentPR?.layers.find(
    (l) => layerKey(l) === selectedLayer
  );

  const isLight = theme === 'light';
  const textMain = isLight ? 'text-nuances-black' : 'text-nuances-50';
  const textSub = isLight ? 'text-primary-600' : 'text-nuances-200';

  const selectPR = (pr: PullRequest) => {
    setSelectedLayer(null);
    updateParam('pr', prKey(pr));
  };

  return (
    <div
      className={`
        flex
        flex-col
        flex-1
        h-screen
        min-w-0
        p-6
        gap-6
        ${isLight ? 'bg-primary-100' : 'bg-nuances-black'}
      `}
    >
      <div className="flex justify-between items-center">
        <h1 className={`text-[32px] font-extrabold leading-[130%] ${textMain}`}>
          Pull Requests
        </h1>
        <Button
          variant={isLight ? 'primary' : 'secondary'}
          isLoading={pullsQuery.isRefetching}
          onClick={() => pullsQuery.refetch()}
        >
          Refresh
        </Button>
      </div>
      <Input
        variant={theme}
        className="w-full"
        placeholder="Search by pull request ID or branch"
        leftIcon={<SearchIcon />}
        value={search}
        onChange={(e) => updateParam('search', e.target.value)}
      />
      <div className="flex flex-row items-center gap-4">
        <span className={`text-base font-semibold ${textMain}`}>
          {filteredPulls.length} pull requests
        </span>
        <span className={`text-base font-medium ${textSub}`}>Filter by</span>
        <RepositoriesDropdown
          variant={theme}
          selectedRepositories={repositoryFilter}
          setSelectedRepositories={(repos) =>
            updateParam('repositories', repos.join(','))
          }
        />
      </div>
      <div className="flex flex-row flex-1 gap-6 min-h-0">
        <div className="flex flex-col w-96 shrink-0 gap-3 overflow-auto">
          {pullsQuery.isLoading && <span className={textSub}>Loading...</span>}
          {pullsQuery.isError && (
            <span className={textSub}>Could not load pull requests.</span>
          )}
          {pullsQuery.isSuccess && filteredPulls.length === 0 && (
            <span className={textSub}>No pull requests found.</span>
          )}
          {filteredPulls.map((pr) => (
            <button
              key={pr.uid}
              onClick={() => selectPR(pr)}
              className={`
                flex
                flex-col
                gap-2
                p-4
                rounded-2xl
                text-left
                cursor-pointer
                outline-2
                ${
                  prKey(pr) === selectedPR
                    ? 'outline-blue-400'
                    : 'outline-transparent'
                }
                ${isLight ? 'bg-nuances-white shadow-light' : 'bg-nuances-400 shadow-dark'}
              `}
            >
              <div className="flex justify-between items-center gap-2">
                <span className={`text-lg font-bold ${textMain}`}>
                  #{pr.id}
                </span>
                <PullRequestStateTag state={pr.state} />
              </div>
              <span className={`text-sm font-medium ${textSub}`}>
                {pr.repository}
              </span>
              <span className={`flex items-center gap-1 text-sm ${textMain}`}>
                <CodeBranchIcon height={16} width={16} />
                {pr.branch} → {pr.base}
              </span>
              <span className={`text-sm ${textSub}`}>
                {pr.layers.length} ephemeral layer
                {pr.layers.length === 1 ? '' : 's'}
              </span>
            </button>
          ))}
        </div>
        <div className="flex flex-col flex-1 min-w-0 gap-4 overflow-auto">
          {!currentPR ? (
            <span className={textSub}>
              Select a pull request to see its ephemeral layers.
            </span>
          ) : (
            <>
              <div className="flex flex-col gap-1">
                <div className="flex items-center gap-4">
                  <h2 className={`text-2xl font-bold ${textMain}`}>
                    Pull request #{currentPR.id}
                  </h2>
                  <PullRequestStateTag state={currentPR.state} />
                </div>
                <span className={`text-sm ${textSub}`}>
                  {currentPR.namespace}/{currentPR.name} ·{' '}
                  {currentPR.repository} · {currentPR.branch} → {currentPR.base}
                </span>
                <span className={`text-sm ${textSub}`}>
                  Last discovered commit:{' '}
                  {currentPR.lastDiscoveredCommit || '-'} · Last commented
                  commit: {currentPR.lastCommentedCommit || '-'}
                </span>
              </div>
              {currentPR.layers.length === 0 && (
                <span className={textSub}>No ephemeral layers yet.</span>
              )}
              {currentPR.layers.map((layer) => (
                <button
                  key={layer.name}
                  onClick={() => setSelectedLayer(layerKey(layer))}
                  className={`
                    flex
                    flex-col
                    gap-2
                    p-4
                    rounded-2xl
                    text-left
                    cursor-pointer
                    outline-2
                    ${
                      layerKey(layer) === selectedLayer
                        ? 'outline-blue-400'
                        : 'outline-transparent'
                    }
                    ${isLight ? 'bg-nuances-white shadow-light' : 'bg-nuances-400 shadow-dark'}
                  `}
                >
                  <div className="flex justify-between items-center gap-2">
                    <span className={`font-bold ${textMain}`}>
                      {layer.name}
                    </span>
                    <div className="flex items-center gap-2">
                      {layer.isRunning && (
                        <span className={`text-sm ${textSub}`}>Running…</span>
                      )}
                      <Tag variant={layer.state} />
                    </div>
                  </div>
                  <span className={`text-sm ${textSub}`}>{layer.path}</span>
                  <span className={`text-sm ${textMain}`}>
                    {layer.lastResult || 'No plan result yet'}
                  </span>
                </button>
              ))}
              {currentLayer &&
                (currentLayer.lastRun.id ? (
                  <LogsTerminal
                    key={`${layerKey(currentLayer)}/${currentLayer.lastRun.id}`}
                    layer={currentLayer}
                    run={currentLayer.lastRun.id}
                    variant={theme}
                    className="h-[50vh] shrink-0"
                  />
                ) : (
                  <span className={textSub}>
                    This layer has no run yet, so there are no logs.
                  </span>
                ))}
            </>
          )}
        </div>
      </div>
    </div>
  );
};

export default Pulls;
