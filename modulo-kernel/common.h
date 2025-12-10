/* common.h */
#ifndef _COMMON_H
#define _COMMON_H

#include <linux/kernel.h>
#include <linux/module.h>
#include <linux/mm.h>
#include <linux/sched/signal.h>
#include <linux/seq_file.h>
#include <linux/proc_fs.h>
#include <linux/uaccess.h>
#include <linux/sched.h>
#include <linux/sched/task.h>
#include <linux/timekeeping.h>
#include <linux/jiffies.h>
#include <linux/slab.h>
#include <linux/vmalloc.h>
#include <linux/ktime.h>

#define PROC_NAME_SYS "sysinfo_so1_202041390"     
#define PROC_NAME_CONT "continfo_so1_202041390"

#define CMDLINE_MAX 512

/* helper: obtener memoria total/free en KB */
static void get_meminfo_kb(unsigned long *total_kb, unsigned long *free_kb)
{
    struct sysinfo si;
    si_meminfo(&si);
    *total_kb = (si.totalram * (unsigned long)si.mem_unit) / 1024;
    *free_kb  = (si.freeram * (unsigned long)si.mem_unit) / 1024;
}

/* helper: vsz (KB) y rss (KB) a partir de mm_struct */
static void get_mem_from_mm(struct mm_struct *mm, unsigned long *vsz_kb, unsigned long *rss_kb)
{
    if (!mm) {
        *vsz_kb = 0;
        *rss_kb = 0;
        return;
    }
    /* total_vm es en páginas */
    *vsz_kb = (mm->total_vm << (PAGE_SHIFT - 10)); /* pages * PAGE_SIZE / 1024 */
    /* get_mm_rss devuelve páginas */
    *rss_kb = (get_mm_rss(mm) << (PAGE_SHIFT - 10));
}

/* helper: leer cmdline de proceso con access_process_vm */
static int read_task_cmdline(struct task_struct *task, char *buf, int bufsize)
{
    struct mm_struct *mm;
    unsigned long arg_start = 0, arg_end = 0;
    int ret = 0;

    rcu_read_lock();
    mm = get_task_mm(task);
    if (!mm) {
        rcu_read_unlock();
        buf[0] = '\0';
        return 0;
    }

    /* En muchos kernels mm->arg_start/arg_end están disponibles */
#if defined(CONFIG_ARM) || defined(CONFIG_X86)
    arg_start = mm->arg_start;
    arg_end   = mm->arg_end;
#else
    /* fallback - intentar acceder a mm->arg_start/arg_end */
    arg_start = mm->arg_start;
    arg_end   = mm->arg_end;
#endif

    if (arg_start && arg_end && arg_end > arg_start) {
        size_t len = min((size_t)(arg_end - arg_start), (size_t)(bufsize - 1));
        /* access_process_vm copia desde el address space del proceso */
        ret = access_process_vm(task, arg_start, buf, len, 0);
        if (ret > 0) {
            /* los argumentos vienen con '\0' entre ellos; sustituimos por espacios */
            int i;
            for (i = 0; i < ret; ++i) if (buf[i] == '\0') buf[i] = ' ';
            buf[ret] = '\0';
        } else {
            buf[0] = '\0';
        }
    } else {
        buf[0] = '\0';
    }

    mmput(mm);
    rcu_read_unlock();
    return ret;
}

/* helper simple para calcular porcentaje (con cuidado en enteros) */
static double percent_of(unsigned long part_kb, unsigned long total_kb)
{
    if (total_kb == 0) return 0.0;
    return (double)part_kb * 100.0 / (double)total_kb;
}

#endif /* _COMMON_H */
