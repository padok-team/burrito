import React from "react";
import { twMerge } from "tailwind-merge";
import {
  createColumnHelper,
  flexRender,
  getCoreRowModel,
  useReactTable,
} from "@tanstack/react-table";

import Tag from "@/components/tags/Tag";
import Chili from "@/assets/illustrations/Chili";
import SyncIcon from "@/assets/icons/SyncIcon";

import { Layer, LayerState } from "@/types/types";

export interface TableProps {
  className?: string;
  variant?: "light" | "dark";
  data: Layer[];
}

const Table: React.FC<TableProps> = ({
  className,
  variant = "light",
  data,
}) => {
  const columnHelper = createColumnHelper<Layer>();

  const columns = [
    columnHelper.accessor("namespace", {
      header: "Namespace",
    }),
    columnHelper.accessor("name", {
      header: "Name",
    }),
    columnHelper.accessor("state", {
      header: "State",
      cell: (state) => getTag(state.getValue()),
    }),
    columnHelper.accessor("repository", {
      header: "Repository",
    }),
    columnHelper.accessor("branch", {
      header: "Branch",
    }),
    columnHelper.accessor("path", {
      header: "Path",
    }),
    columnHelper.accessor("lastResult", {
      header: "Last result",
      cell: (result) => (
        <div className="relative flex items-center">
          <span>{result.getValue()}</span>
          {result.row.original.isRunning && (
            <div
              className={`absolute
              -right-5
              flex
              items-center
              justify-end
              h-[calc(100%_+_24px)]
              min-w-full
              w-full
              rounded-xl
              pr-4
              pointer-events-none
              ${
                variant === "light"
                  ? "bg-[linear-gradient(270deg,_#FFF_56.84%,_rgba(255,_255,_255,_0.00)_100%)]"
                  : "bg-[linear-gradient(270deg,_#000_56.84%,_rgba(0,_0,_0,_0.00)_100%)]"
              }`}
            >
              <div className="flex items-center gap-2 text-blue-500 fill-blue-500">
                <span className="text-sm font-semibold">Running</span>
                <SyncIcon height={16} width={16} />
              </div>
            </div>
          )}
        </div>
      ),
    }),
  ];

  const getTag = (state: LayerState) => {
    return (
      <div className="relative flex items-center">
        <Tag variant={state} />
        {state === "error" && (
          <Chili
            className="absolute translate-x-16 rotate-[-21deg]"
            height={24}
            width={24}
          />
        )}
      </div>
    );
  };

  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
  });

  const styles = {
    header: {
      light: `text-primary-600`,
      dark: `text-nuances-300`,
    },
    row: {
      base: {
        light: `text-nuances-black
          hover:bg-nuances-white
          hover:shadow-light`, // BUG: not working on Safari
        dark: `text-nuances-50
          hover:bg-nuances-400
          hover:shadow-dark`, // BUG: not working on Safari
      },
      running: {
        light: `outline-blue-400`,
        dark: `outline-blue-500`,
      },
    },
  };

  return (
    <table className={twMerge(`w-full border-collapse`, className)}>
      <thead>
        {table.getHeaderGroups().map((headerGroup) => (
          <tr key={headerGroup.id} className={`${styles.header[variant]}`}>
            {headerGroup.headers.map((header) => (
              <th
                key={header.id}
                className={`text-left
                  text-base
                  font-normal
                  px-6
                  pb-4`}
              >
                {header.isPlaceholder
                  ? null
                  : flexRender(
                      header.column.columnDef.header,
                      header.getContext()
                    )}
              </th>
            ))}
          </tr>
        ))}
      </thead>
      <tbody>
        {table.getRowModel().rows.map((row) => (
          <tr
            key={row.id}
            className={twMerge(
              `${styles.row.base[variant]}`,
              row.original.isRunning &&
                `rounded-2xl
                outline
                outline-4
                -outline-offset-4
                ${styles.row.running[variant]}`
            )}
          >
            {row.getVisibleCells().map((cell) => (
              <td
                key={cell.id}
                className={`text-left
                  text-base
                  font-semibold
                  px-6
                  py-4
                  ${
                    cell.row.original.isRunning &&
                    "first:rounded-l-2xl last:rounded-r-2xl"
                  }`}
              >
                {flexRender(cell.column.columnDef.cell, cell.getContext())}
              </td>
            ))}
          </tr>
        ))}
      </tbody>
    </table>
  );
};

export default Table;
