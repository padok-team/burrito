import React, { useState } from "react";
import { twMerge } from "tailwind-merge";
import {
  createColumnHelper,
  flexRender,
  getCoreRowModel,
  useReactTable,
} from "@tanstack/react-table";
import { Tooltip } from "react-tooltip";

import TableLoader from "@/components/loaders/TableLoader";
import LogsButton from "@/components/buttons/LogsButton";
import Running from "@/components/widgets/Running";
import Tag from "@/components/widgets/Tag";
import ChiliLight from "@/assets/illustrations/ChiliLight";
import ChiliDark from "@/assets/illustrations/ChiliDark";
import CodeBranchIcon from "@/assets/icons/CodeBranchIcon";

import { Layer, LayerState } from "@/clients/layers/types";

export interface TableProps {
  className?: string;
  variant?: "light" | "dark";
  isLoading?: boolean;
  data: Layer[];
}

const Table: React.FC<TableProps> = ({
  className,
  variant = "light",
  isLoading,
  data,
}) => {
  const columnHelper = createColumnHelper<Layer>();
  const [hoveredRow, setHoveredRow] = useState<Layer | null>(null);

  const columns = [
    columnHelper.accessor("isPR", {
      header: "",
      cell: (isPR) => isPR.getValue() && <CodeBranchIcon className="-mr-6" />,
    }),
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
        <div className="relative flex items-center h-full">
          <span>{result.getValue()}</span>
          {result.row.original === hoveredRow ? (
            <div
              className={`
                absolute
                -right-5
                flex
                items-center
                justify-end
                h-[calc(100%_+_25px)]
                min-w-full
                w-full
                pr-4
                ${result.row.original.isRunning && "rounded-xl"}
                ${
                  variant === "light"
                    ? "bg-[linear-gradient(270deg,_#FFF_58.7%,_rgba(255,_255,_255,_0.00)_100%)]"
                    : "bg-[linear-gradient(270deg,_#252525_58.7%,_rgba(37,_37,_37,_0.00)_100%)]"
                }
              `}
            >
              <LogsButton variant={variant} />
            </div>
          ) : result.row.original.isRunning ? (
            <div
              className={`
                absolute
                -right-5
                flex
                items-center
                justify-end
                h-[calc(100%_+_25px)]
                min-w-full
                w-full
                rounded-xl
                pr-4
                pointer-events-none
                ${
                  variant === "light"
                    ? "bg-[linear-gradient(270deg,_#FFF_58.7%,_rgba(255,_255,_255,_0.00)_100%)]"
                    : "bg-[linear-gradient(270deg,_#000_58.7%,_rgba(0,_0,_0,_0.00)_100%)]"
                }
              `}
            >
              <Running />
            </div>
          ) : null}
        </div>
      ),
    }),
  ];

  const getTag = (state: LayerState) => {
    return (
      <div className="relative flex items-center">
        <Tag variant={state} />
        {state === "error" &&
          (variant === "light" ? (
            <ChiliLight
              className="absolute translate-x-16 rotate-[-21deg]"
              height={24}
              width={24}
            />
          ) : (
            <ChiliDark
              className="absolute translate-x-16 rotate-[-21deg]"
              height={24}
              width={24}
            />
          ))}
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
          fill-nuances-black
          hover:bg-nuances-white
          hover:shadow-light`, // BUG: not working on Safari
        dark: `text-nuances-50
          fill-nuances-50
          hover:bg-nuances-400
          hover:shadow-dark`, // BUG: not working on Safari
      },
      running: {
        light: `outline-blue-400`,
        dark: `outline-blue-500`,
      },
    },
    separator: {
      light: `border-primary-500`,
      dark: `border-nuances-300`,
    },
  };

  return (
    <div>
      <table className={twMerge(`w-full border-collapse h-[1px]`, className)}>
        {/* HACK: 1px height actually ignored but required to make cell div full size */}
        <thead>
          {table.getHeaderGroups().map((headerGroup) => (
            <tr key={headerGroup.id} className={`${styles.header[variant]}`}>
              {headerGroup.headers.map((header, index) => (
                <th
                  key={header.id}
                  className={`
                    relative
                    text-left
                    text-base
                    font-normal
                    px-6
                    pb-4
                  `}
                >
                  {header.isPlaceholder
                    ? null
                    : flexRender(
                        header.column.columnDef.header,
                        header.getContext()
                      )}
                  {index === 0 ? (
                    <hr
                      className={`
                        absolute
                        right-0
                        bottom-0
                        w-[calc(100%_-_25px)]
                        ${styles.separator[variant]}
                      `}
                    />
                  ) : index === headerGroup.headers.length - 1 ? (
                    <hr
                      className={`
                        absolute
                        left-0
                        bottom-0
                        w-[calc(100%_-_25px)]
                        ${styles.separator[variant]}
                      `}
                    />
                  ) : (
                    <hr
                      className={`
                        absolute
                        bottom-0
                        left-0
                        w-full
                        ${styles.separator[variant]}
                      `}
                    />
                  )}
                </th>
              ))}
            </tr>
          ))}
        </thead>
        <tbody>
          {isLoading
            ? Array.from({ length: 100 }).map((_, index) => (
                <tr
                  key={index}
                  className={twMerge(
                    `h-full
                    ${styles.row.base[variant]}`
                  )}
                >
                  {table.getAllColumns().map((_, index) => (
                    <td
                      key={index}
                      className={`relative
                        text-left
                        h-full
                        text-base
                        font-semibold
                        px-6
                        py-4`}
                    >
                      <TableLoader variant={variant} />
                      {index === 0 ? (
                        <hr
                          className={`
                            absolute
                            right-0
                            bottom-0
                            w-[calc(100%_-_25px)]
                            ${styles.separator[variant]}
                          `}
                        />
                      ) : index === table.getAllColumns().length - 1 ? (
                        <hr
                          className={`
                            absolute
                            left-0
                            bottom-0
                            w-[calc(100%_-_25px)]
                            ${styles.separator[variant]}
                          `}
                        />
                      ) : (
                        <hr
                          className={`
                            absolute
                            bottom-0
                            left-0
                            w-full
                            ${styles.separator[variant]}
                          `}
                        />
                      )}
                    </td>
                  ))}
                </tr>
              ))
            : table.getRowModel().rows.map((row) => (
                <tr
                  key={row.id}
                  className={twMerge(
                    `h-full
                  ${styles.row.base[variant]}`,
                    row.original.isRunning &&
                      `rounded-2xl
                      outline
                      outline-4
                      -outline-offset-4
                      ${styles.row.running[variant]}`
                  )}
                  onMouseEnter={() => setHoveredRow(row.original)}
                  onMouseLeave={() => setHoveredRow(null)}
                >
                  {row.getVisibleCells().map((cell, index) => (
                    <td
                      key={cell.id}
                      className={twMerge(
                        `relative
                        text-left
                        h-full
                        text-base
                        font-semibold
                        px-6
                        py-4`,
                        cell.row.original.isRunning &&
                          "first:rounded-l-2xl last:rounded-r-2xl"
                      )}
                      data-tooltip-id="table-tooltip"
                      data-tooltip-content={
                        cell.column.id === "lastResult" &&
                        cell.row.original.isRunning
                          ? (cell.getValue() as string)
                          : null
                      }
                    >
                      {flexRender(
                        cell.column.columnDef.cell,
                        cell.getContext()
                      )}
                      {index === 0 ? (
                        <hr
                          className={`
                            absolute
                            right-0
                            bottom-0
                            w-[calc(100%_-_25px)]
                            ${styles.separator[variant]}
                          `}
                        />
                      ) : index === row.getVisibleCells().length - 1 ? (
                        <hr
                          className={`
                            absolute
                            left-0
                            bottom-0
                            w-[calc(100%_-_25px)]
                            ${styles.separator[variant]}
                          `}
                        />
                      ) : (
                        <hr
                          className={`
                            absolute
                            bottom-0
                            left-0
                            w-full
                            ${styles.separator[variant]}
                          `}
                        />
                      )}
                    </td>
                  ))}
                </tr>
              ))}
        </tbody>
      </table>
      <Tooltip
        opacity={1}
        id="table-tooltip"
        variant={variant === "light" ? "dark" : "light"}
      />
    </div>
  );
};

export default Table;
