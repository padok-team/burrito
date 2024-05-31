import { useCallback, useEffect } from 'react'

function useFixMissingScroll({ hasMoreItems, fetchMoreItems, mainElement }: { hasMoreItems: boolean, fetchMoreItems: () => void, mainElement: HTMLElement }) {

  const fetchCb = useCallback(() => {
    fetchMoreItems()
  }, [fetchMoreItems])

  useEffect(() => {
    const hasScroll = mainElement ? mainElement.scrollHeight > mainElement.clientHeight : false
    if (!hasScroll && hasMoreItems) {
      setTimeout(() => {
        fetchCb()
      }, 100)
    }
  }, [hasMoreItems, fetchCb, mainElement])
}

export default useFixMissingScroll
