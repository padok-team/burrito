import React from "react";
import { twMerge } from "tailwind-merge";
import {
  createColumnHelper,
  flexRender,
  getCoreRowModel,
  useReactTable,
} from "@tanstack/react-table";

import Tag from "@/components/tags/Tag";

import { Layer } from "@/types/types";

export interface TableProps {
  className?: string;
  variant?: "light" | "dark";
  data: Layer[];
}

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
    cell: (status) => (
      <div className="flex">
        <Tag variant={status.getValue()} />
      </div>
    ),
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
  }),
];

const Table: React.FC<TableProps> = ({
  className,
  variant = "light",
  data,
}) => {
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
      light: `text-nuances-black`,
      dark: `text-nuances-50`,
    },
  };

  return (
    <table className={twMerge(`w-full border-collapse`, className)}>
      <thead>
        {table.getHeaderGroups().map((headerGroup) => (
          <tr key={headerGroup.id}>
            {headerGroup.headers.map((header) => (
              <th
                key={header.id}
                className={`text-left
                  text-base
                  font-normal
                  px-6
                  pb-4
                  ${styles.header[variant]}`}
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
          <tr key={row.id}>
            {row.getVisibleCells().map((cell) => (
              <td
                key={cell.id}
                className={`text-left
                  text-base
                  font-semibold
                  px-6
                  py-4
                  ${styles.row[variant]}`}
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
