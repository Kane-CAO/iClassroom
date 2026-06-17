import { X } from 'lucide-react'

export interface PreviewImage {
  id?: string
  /** 图片地址（兼容原型的 src / url 两种字段）。 */
  src: string
  /** 标题 / 文件名。 */
  label?: string
}

interface ImagePreviewProps {
  images: PreviewImage[]
  columns?: 2 | 3
  /** 是否显示标题脚注（讲师端提交图片带 caption）。 */
  showCaption?: boolean
  /** 传入则在每张图上显示删除按钮（学生上传预览）。注意：本阶段不实现真实上传。 */
  onRemove?: (id: string) => void
  emptyText?: string
  imgHeightClass?: string
}

// 图片预览网格。复用于：学生答题图片预览、讲师批改提交图片。
// 迁移自 student.html 的 renderPreviews() 与 iClassroom.html 的 Uploaded Images。
export default function ImagePreview({
  images,
  columns = 3,
  showCaption = false,
  onRemove,
  emptyText = '本题暂无已上传图片。',
  imgHeightClass = 'h-40',
}: ImagePreviewProps) {
  const colClass = columns === 2 ? 'grid-cols-2' : 'grid-cols-3'

  if (images.length === 0) {
    return (
      <div
        className={`grid ${colClass} gap-3`}
      >
        <div className="col-span-full rounded-lg border border-dashed border-line p-6 text-center text-sm text-muted dark:border-slate-800 dark:text-slate-400">
          {emptyText}
        </div>
      </div>
    )
  }

  return (
    <div className={`grid ${colClass} gap-3`}>
      {images.map((image, index) => (
        <figure
          key={image.id ?? `${image.src}-${index}`}
          className="hover-zoom relative overflow-hidden rounded-lg border border-line bg-slate-100 dark:border-slate-800 dark:bg-slate-950"
        >
          <img className={`${imgHeightClass} w-full object-cover`} src={image.src} alt={image.label ?? '图片预览'} />
          {showCaption && image.label && (
            <figcaption className="border-t border-line px-3 py-2 text-xs font-semibold dark:border-slate-800">
              {image.label}
            </figcaption>
          )}
          {onRemove && image.id && (
            <button
              onClick={() => onRemove(image.id!)}
              aria-label="移除图片"
              className="absolute right-1.5 top-1.5 rounded-full bg-white/90 p-1 text-slate-600 shadow-sm hover:text-rose-600 dark:bg-slate-900/90 dark:text-slate-300"
            >
              <X className="h-3.5 w-3.5" />
            </button>
          )}
        </figure>
      ))}
    </div>
  )
}
