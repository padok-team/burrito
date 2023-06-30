import axios, { AxiosRequestConfig, AxiosResponse } from 'axios';

const config = async () => {
  return {};
};

export const get = async <ResponseT>(
  url: string,
  extraConfig?: AxiosRequestConfig,
): Promise<AxiosResponse<ResponseT>> => {
  return axios.get(url, { ...(await config()), ...extraConfig });
};

export const post = async <ResponseT, DataT = void>(
  url: string,
  data?: DataT,
): Promise<AxiosResponse<ResponseT>> => {
  return axios.post(url, data, await config());
};

export const put = async <ResponseT, DataT = void>(
  url: string,
  data?: DataT,
): Promise<AxiosResponse<ResponseT>> => {
  return axios.put(url, data, await config());
};

export const patch = async <ResponseT, DataT = void>(
  url: string,
  data?: DataT,
): Promise<AxiosResponse<ResponseT>> => {
  return axios.patch(url, data, await config());
};

export const deleteRequest = async (
  url: string,
): Promise<AxiosResponse<void>> => {
  return axios.delete(url, await config());
};
